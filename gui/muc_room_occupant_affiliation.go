package gui

import (
	"github.com/coyim/coyim/i18n"
	"github.com/coyim/coyim/session"
	"github.com/coyim/coyim/session/muc"
	"github.com/coyim/coyim/session/muc/data"
	"github.com/coyim/coyim/xmpp/jid"
	"github.com/coyim/gotk3adapter/gtki"
	log "github.com/sirupsen/logrus"
)

func (r *roomViewRosterInfo) onChangeAffiliation() {
	l := r.log.WithFields(log.Fields{
		"occupant": r.occupant.RealJid,
		"where":    "occupantAffiliationUpdate",
	})

	av := r.newOccupantAffiliationUpdateView(r.account, r.roomID, r.occupant, func(a data.Affiliation) {
		l.WithField("affiliation", a.Name()).Info("The occupant affiliation has been updated")
		doInUIThread(r.refresh)
	})

	av.show()
}

type occupantAffiliationUpdateView struct {
	account              *account
	roomID               jid.Bare
	occupant             *muc.Occupant
	cancel               chan bool
	onAffiliationUpdated func(data.Affiliation)

	dialog            gtki.Dialog      `gtk-widget:"affiliation-dialog"`
	affiliationLabel  gtki.Label       `gtk-widget:"affiliation-type-label"`
	adminRadio        gtki.RadioButton `gtk-widget:"affiliation-admin"`
	noneRadio         gtki.RadioButton `gtk-widget:"affiliation-none"`
	reasonLabel       gtki.Label       `gtk-widget:"affiliation-reason-label"`
	reasonEntry       gtki.TextView    `gtk-widget:"affiliation-reason-entry"`
	applyButton       gtki.Button      `gtk-widget:"affiliation-apply-button"`
	notificationsArea gtki.Box         `gtk-widget:"notifications-area"`
	spinnerArea       gtki.Box         `gtk-widget:"spinner-area"`

	notifications *notifications
	spinner       *spinner
}

func (r *roomViewRosterInfo) newOccupantAffiliationUpdateView(a *account, roomID jid.Bare, o *muc.Occupant, onAffiliationUpdated func(data.Affiliation)) *occupantAffiliationUpdateView {
	av := &occupantAffiliationUpdateView{
		account:              a,
		roomID:               roomID,
		occupant:             o,
		onAffiliationUpdated: onAffiliationUpdated,
	}

	av.initBuilder()
	av.initNotificationsAndSpinner(r.u)
	av.initDefaults()

	return av
}

func (av *occupantAffiliationUpdateView) initBuilder() {
	builder := newBuilder("MUCRoomAffiliationDialog")
	panicOnDevError(builder.bindObjects(av))

	builder.ConnectSignals(map[string]interface{}{
		"on_cancel": av.onCancel,
		"on_apply":  av.onApply,
	})
}

func (av *occupantAffiliationUpdateView) initNotificationsAndSpinner(u *gtkUI) {
	av.notifications = u.newNotificationsComponent()
	av.spinner = u.newSpinnerComponent()

	av.notificationsArea.Add(av.notifications.widget())
	av.spinnerArea.Add(av.spinner.widget())
}

func (av *occupantAffiliationUpdateView) initDefaults() {
	mucStyles.setFormSectionLabelStyle(av.affiliationLabel)

	av.deactivateRadioButton(av.adminRadio)
	av.deactivateRadioButton(av.noneRadio)

	switch av.occupant.Affiliation.(type) {
	case *data.AdminAffiliation:
		av.activateRadioButton(av.adminRadio)
	case *data.NoneAffiliation:
		av.activateRadioButton(av.noneRadio)
	}
}

// activateRadioButton MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) activateRadioButton(r gtki.RadioButton) {
	r.SetActive(true)
}

// deactivateRadioButton MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) deactivateRadioButton(r gtki.RadioButton) {
	r.SetActive(false)
}

// disableAffiliationRadios MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) disableAffiliationRadios() {
	disableField(av.adminRadio)
	disableField(av.noneRadio)
}

// enableAffiliationRadios MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) enableAffiliationRadios() {
	enableField(av.adminRadio)
	enableField(av.noneRadio)
}

// disableFieldsAndShowSpinner MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) disableFieldsAndShowSpinner() {
	av.disableAffiliationRadios()
	av.applyButton.SetSensitive(false)
	av.spinner.show()
}

// enableFieldsAndHideSpinner MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) enableFieldsAndHideSpinner() {
	av.enableAffiliationRadios()
	av.applyButton.SetSensitive(true)
	av.spinner.hide()
}

// onCancel MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) onCancel() {
	go func() {
		if av.cancel != nil {
			av.cancel <- true
			av.cancel = nil
		}
	}()

	av.close()
}

// onApply MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) onApply() {
	previousAffiliation := av.occupant.Affiliation

	switch {
	case av.adminRadio.GetActive():
		av.occupant.ChangeAffiliationToAdmin()
	case av.noneRadio.GetActive():
		av.occupant.ChangeAffiliationToNone()
	}

	av.disableFieldsAndShowSpinner()

	go av.updateOccupantAffiliation(previousAffiliation)
}

// updateOccupantAffiliation MUST NOT be called from the UI thread
func (av *occupantAffiliationUpdateView) updateOccupantAffiliation(previousAffiliation data.Affiliation) {
	av.cancel = make(chan bool)
	sc, ec := av.account.session.UpdateOccupantAffiliation(av.roomID, av.occupant, getTextViewText(av.reasonEntry))

	select {
	case <-sc:
		av.onAffiliationUpdateFinished()
	case err := <-ec:
		av.onAffiliationUpdateError(previousAffiliation, err)
	case <-av.cancel:
		// TODO: should we update the affiliation to its previous value?
	}
}

// onAffiliationUpdatedFinished MUST NOT be called from the UI thread
func (av *occupantAffiliationUpdateView) onAffiliationUpdateFinished() {
	av.onAffiliationUpdated(av.occupant.Affiliation)
	doInUIThread(av.close)
}

// onAffiliationUpdateError MUST NOT be called from the UI thread
func (av *occupantAffiliationUpdateView) onAffiliationUpdateError(previousAffiliation data.Affiliation, err error) {
	av.occupant.UpdateAffiliation(previousAffiliation)
	doInUIThread(func() {
		av.enableFieldsAndHideSpinner()
		av.notifications.error(affiliationUpdateErrorMessage(err))
	})
}

// show MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) show() {
	av.dialog.Show()
}

// close MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) close() {
	av.dialog.Destroy()
}

func occupantAffiliationName(a data.Affiliation) string {
	switch a.(type) {
	case *data.OwnerAffiliation:
		return i18n.Local("Owner")
	case *data.AdminAffiliation:
		return i18n.Local("Admin")
	case *data.MemberAffiliation:
		return i18n.Local("Member")
	case *data.OutcastAffiliation:
		return i18n.Local("Outcast")
	case *data.NoneAffiliation:
		return i18n.Local("None")
	default:
		// This should not be possible but we need it to not complain with golang
		return ""
	}
}

func affiliationUpdateErrorMessage(err error) string {
	switch err {
	case session.ErrUpdateOccupantAffiliationResponse:
		return i18n.Local("We couldn't update the occupant affiliation because either you don't have permissions to do it or the server is busy. Please tray again.")
	default:
		return i18n.Local("An error occurred when updating the occupant affiliation. Please try again.")
	}
}