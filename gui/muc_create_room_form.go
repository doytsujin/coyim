package gui

import (
	"github.com/coyim/coyim/coylog"
	"github.com/coyim/coyim/i18n"
	"github.com/coyim/coyim/xmpp/jid"
	"github.com/coyim/gotk3adapter/gtki"
	"github.com/golang-collections/collections/set"
)

type mucCreateRoomViewForm struct {
	isShown               bool
	builder               *builder
	accountsComponent     *connectedAccountsComponent
	chatServicesComponent *chatServicesComponent

	view             gtki.Box         `gtk-widget:"create-room-form"`
	roomEntry        gtki.Entry       `gtk-widget:"room-name-entry"`
	roomAutoJoin     gtki.CheckButton `gtk-widget:"autojoin-check-button"`
	createButton     gtki.Button      `gtk-widget:"create-room-button"`
	spinnerBox       gtki.Box         `gtk-widget:"spinner-box"`
	notificationArea gtki.Box         `gtk-widget:"notification-area-box"`

	spinner       *spinner
	notifications *notifications

	roomNameConflictList    *set.Set
	createRoom              func(*account, jid.Bare)
	onCheckFieldsConditions func(string, string, *account) bool

	log func(*account, jid.Bare) coylog.Logger
}

func (v *mucCreateRoomView) newCreateRoomForm() *mucCreateRoomViewForm {
	f := &mucCreateRoomViewForm{
		roomNameConflictList: set.New(),
		log:                  v.log,
	}

	f.initBuilder(v)
	f.initNotifications(v)
	f.initChatServices(v)
	f.initConnectedAccounts(v)
	f.initDefaults(v)

	return f
}

func (f *mucCreateRoomViewForm) initBuilder(v *mucCreateRoomView) {
	f.builder = newBuilder("MUCCreateRoomForm")
	panicOnDevError(f.builder.bindObjects(f))

	f.builder.ConnectSignals(map[string]interface{}{
		"on_cancel":          v.onCancel,
		"on_create_room":     f.onCreateRoom,
		"on_roomName_change": f.enableCreationIfConditionsAreMet,
		"on_roomAutoJoin_toggled": func() {
			v.updateAutoJoinValue(f.roomAutoJoin.GetActive())
		},
		"on_chatServiceEntry_change": f.enableCreationIfConditionsAreMet,
	})
}

func (f *mucCreateRoomViewForm) initNotifications(v *mucCreateRoomView) {
	f.notifications = v.u.newNotifications(f.notificationArea)
}

func (f *mucCreateRoomViewForm) initChatServices(v *mucCreateRoomView) {
	chatServicesList := f.builder.get("chat-services-list").(gtki.ComboBoxText)
	chatServicesEntry := f.builder.get("chat-services-entry").(gtki.Entry)
	f.chatServicesComponent = v.u.createChatServicesComponent(chatServicesList, chatServicesEntry, f.enableCreationIfConditionsAreMet)
}

func (f *mucCreateRoomViewForm) initConnectedAccounts(v *mucCreateRoomView) {
	account := f.builder.get("accounts").(gtki.ComboBox)
	f.accountsComponent = v.u.createConnectedAccountsComponent(account, f.notifications, f.updateServicesBasedOnAccount, f.onNoAccountsConnected)
}

func (f *mucCreateRoomViewForm) initDefaults(v *mucCreateRoomView) {
	f.spinner = newSpinner()
	f.spinnerBox.Add(f.spinner.getWidget())
}

func (v *mucCreateRoomView) initCreateRoomForm() *mucCreateRoomViewForm {
	f := v.newCreateRoomForm()

	f.createRoom = func(ca *account, roomID jid.Bare) {
		errors := make(chan error)
		v.createRoom(ca, roomID, errors)
		go f.listenToCreateError(roomID, errors)
	}

	f.addCallbacks(v)

	return f
}

func (f *mucCreateRoomViewForm) listenToCreateError(roomID jid.Bare, errors chan error) {
	err := <-errors

	switch err {
	case errCreateRoomCheckIfExistsFails:
		doInUIThread(f.onCreateRoomCheckIfExistsFails)

	case errCreateRoomAlreadyExists:
		f.roomNameConflictList.Insert(roomID.String())
		doInUIThread(f.onCreateRoomAlreadyExists)

	case errCreateRoomFailed:
		doInUIThread(func() {
			f.onCreateRoomFailed(err)
		})
	}
}

func (f *mucCreateRoomViewForm) onCreateRoomCheckIfExistsFails() {
	f.notifications.error(i18n.Local("Couldn't connect to the service, please verify that it exists or try again later."))
	f.spinner.hide()
	f.enableFields()
}

func (f *mucCreateRoomViewForm) onCreateRoomAlreadyExists() {
	f.notifications.error(i18n.Local("That room already exists, try again with a different name."))
	f.spinner.hide()
	f.enableFields()
	f.createButton.SetSensitive(false)
}

func (f *mucCreateRoomViewForm) onCreateRoomFailed(err error) {
	displayErr, ok := supportedCreateMUCErrors[err]
	if ok {
		f.notifications.error(displayErr)
	} else {
		f.notifications.error(i18n.Local("Could not create the room."))
	}
}

func (f *mucCreateRoomViewForm) addCallbacks(v *mucCreateRoomView) {
	v.onAutoJoin.add(func() {
		f.onAutoJoinChange(v.autoJoin)
	})

	v.onDestroy.add(f.destroy)
}

func (f *mucCreateRoomViewForm) showCreateForm(v *mucCreateRoomView) {
	v.success.reset()
	v.container.Remove(v.success.view)
	f.reset()
	v.container.Add(f.view)
	f.isShown = true
}

func (f *mucCreateRoomViewForm) onAutoJoinChange(v bool) {
	if v {
		f.createButton.SetProperty("label", i18n.Local("Create Room & Join"))
	} else {
		f.createButton.SetProperty("label", i18n.Local("Create Room"))
	}
}

func (f *mucCreateRoomViewForm) onCreateRoom() {
	f.notifications.clearErrors()

	roomName, _ := f.roomEntry.GetText()
	local := jid.NewLocal(roomName)
	if !local.Valid() {
		f.log(nil, nil).WithField("local", roomName).Error("Trying to create a room with an invalid local")
		f.notifications.error(i18n.Local("You must provide a valid room name."))
		return
	}

	domain := f.chatServicesComponent.currentService()
	if !domain.Valid() {
		f.log(nil, nil).WithField("domain", domain).Error("Trying to create a room with an invalid domain")
		f.notifications.error(i18n.Local("You must provide a valid service name."))
		return
	}

	roomID := jid.NewBare(local, domain)

	ca := f.accountsComponent.currentAccount()
	if ca == nil {
		f.log(nil, roomID).Error("No account was selected to create the room")
		f.notifications.error(i18n.Local("No account is selected, please select one account from the list or connect to one."))
		return
	}

	f.beforeCreatingTheRoom()

	go f.createRoom(ca, roomID)
}

func (f *mucCreateRoomViewForm) beforeCreatingTheRoom() {
	f.spinner.show()
	f.disableFields()
}

func (f *mucCreateRoomViewForm) destroy() {
	f.isShown = false

	if f.accountsComponent != nil {
		f.accountsComponent.onDestroy()
	}
}

func (f *mucCreateRoomViewForm) clearFields() {
	f.roomEntry.SetText("")
	f.enableCreationIfConditionsAreMet()
}

func (f *mucCreateRoomViewForm) reset() {
	f.spinner.hide()
	f.enableFields()
	f.clearFields()
}

func (f *mucCreateRoomViewForm) setFieldsSensitive(v bool) {
	f.createButton.SetSensitive(v)
	f.roomEntry.SetSensitive(v)
	f.roomAutoJoin.SetSensitive(v)
}

func (f *mucCreateRoomViewForm) disableFields() {
	f.setFieldsSensitive(false)
	f.accountsComponent.disableAccountInput()
	f.chatServicesComponent.disableServiceInput()
}

func (f *mucCreateRoomViewForm) enableFields() {
	f.setFieldsSensitive(true)
	f.accountsComponent.enableAccountInput()
	f.chatServicesComponent.enableServiceInput()
}

func (f *mucCreateRoomViewForm) updateServicesBasedOnAccount(ca *account) {
	doInUIThread(func() {
		f.notifications.clearErrors()
		f.enableCreationIfConditionsAreMet()
	})
	go f.chatServicesComponent.updateServicesBasedOnAccount(ca)
}

func (f *mucCreateRoomViewForm) onNoAccountsConnected() {
	doInUIThread(func() {
		f.enableCreationIfConditionsAreMet()
		f.chatServicesComponent.removeAll()
	})
}

func (f *mucCreateRoomViewForm) enableCreationIfConditionsAreMet() {
	var canCreate bool
	var message string

	defer func() {
		f.createButton.SetSensitive(canCreate)
		if message != "" {
			f.notifications.error(message)
		}
	}()

	// Let the connected accounts component show any errors if it have one
	if f.accountsComponent.hasAccounts() {
		f.notifications.clearErrors()
	}

	validations := []func() (ok bool, message string){
		f.validateRoomRequiredFields,
		f.validateIfRoomIsInConflict,
	}

	for _, v := range validations {
		canCreate, message = v()
		if !canCreate {
			return
		}
	}
}

func (f *mucCreateRoomViewForm) validateRoomRequiredFields() (bool, string) {
	v := f.roomNameFromEntry() != "" && f.chatServicesComponent.hasServiceValue() && f.accountsComponent.currentAccount() != nil
	return v, ""
}

func (f *mucCreateRoomViewForm) validateIfRoomIsInConflict() (bool, string) {
	if !f.isRoomInConflict(f.roomNameFromEntry(), f.chatServicesComponent.currentServiceValue()) {
		return false, i18n.Local("That room already exists, try again with a different name.")
	}
	return true, ""
}

func (f *mucCreateRoomViewForm) isRoomInConflict(local, domain string) bool {
	roomID := jid.NewBareFromStrings(local, domain)
	return f.roomNameConflictList.Has(roomID.String())
}

func (f *mucCreateRoomViewForm) roomNameFromEntry() string {
	t, _ := f.roomEntry.GetText()
	return t
}

func setEnabled(w gtki.Widget, enable bool) {
	w.SetSensitive(enable)
}
