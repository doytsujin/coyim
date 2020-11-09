package gui

import (
	"github.com/coyim/coyim/coylog"
	"github.com/coyim/coyim/i18n"

	"github.com/coyim/coyim/session/muc"
	"github.com/coyim/coyim/session/muc/data"
	"github.com/coyim/coyim/xmpp/jid"
	"github.com/coyim/gotk3adapter/gtki"
)

type roomView struct {
	u       *gtkUI
	account *account
	builder *builder

	room *muc.Room

	cancel chan bool

	opened   bool
	returnTo func()

	window                 gtki.Window  `gtk-widget:"roomWindow"`
	content                gtki.Box     `gtk-widget:"boxMainView"`
	messagesOverlay        gtki.Overlay `gtk-widget:"messagesOverlay"`
	messagesOverlayBox     gtki.Box     `gtk-widget:"messagesOverlayBox"`
	messagesBox            gtki.Box     `gtk-widget:"messagesBox"`
	notificationBox        gtki.Box     `gtk-widget:"notificationBox"`
	loadingNotificationBox gtki.Box     `gtk-widget:"loadingNotificationBox"`

	spinner         *spinner
	notifications   *notifications
	loadingInfoBar  *roomViewLoadingInfoBar
	warnings        *roomViewWarningsOverlay
	warningsInfoBar *roomViewWarningsInfoBar

	subscribers *roomViewSubscribers

	main    *roomViewMain
	menu    *roomViewMenu
	toolbar *roomViewToolbar
	roster  *roomViewRoster
	conv    *roomViewConversation
	lobby   *roomViewLobby

	log coylog.Logger
}

func newRoomView(u *gtkUI, a *account, roomID jid.Bare) *roomView {
	view := &roomView{
		u:       u,
		account: a,
	}

	// TODO: We already know this need to change
	view.room = a.newRoomModel(roomID)
	view.log = a.log.WithField("room", roomID)

	view.room.Subscribe(view.handleRoomEvent)

	view.subscribers = newRoomViewSubscribers(view.roomID(), view.log)

	view.initBuilderAndSignals()
	view.initDefaults()
	view.initRoomMenu()
	view.initSubscribers()

	view.toolbar = view.newRoomViewToolbar()
	view.roster = view.newRoomViewRoster()
	view.conv = view.newRoomViewConversation()

	view.spinner = newSpinner()
	view.notifications = u.newNotifications(view.notificationBox)
	view.loadingInfoBar = view.newRoomViewLoadingInfoBar(view.loadingNotificationBox)
	view.warnings = view.newRoomViewWarningsOverlay(view.closeNotificationsOverlay)
	view.warningsInfoBar = view.newRoomViewWarningsInfoBar(view.showWarnings, view.removeWarningsInfobar)

	go view.requestRoomInfo()

	return view
}

func (v *roomView) initBuilderAndSignals() {
	v.builder = newBuilder("MUCRoomWindow")

	panicOnDevError(v.builder.bindObjects(v))

	v.builder.ConnectSignals(map[string]interface{}{
		"on_destroy_window": v.onDestroyWindow,
	})
}

func (v *roomView) initDefaults() {
	v.setTitle(i18n.Localf("%s [%s]", v.roomID(), v.account.Account()))
}

func (v *roomView) initSubscribers() {
	v.subscribe("room", func(ev roomViewEvent) {
		doInUIThread(func() {
			v.onEventReceived(ev)
		})
	})
}

func (v *roomView) onEventReceived(ev roomViewEvent) {
	switch t := ev.(type) {
	case roomConfigReceivedEvent:
		v.roomConfigReceivedEvent(t.config)
	case roomConfigRequestTimeoutEvent:
		v.roomConfigRequestTimeoutEvent()
	}

	// Actually, we prefer to always refresh the room menu every time an event
	// is received, because in that way we can have the right menu printed based
	// on the latest configuration of the room
	v.refreshRoomMenu()
}

func (v *roomView) requestRoomInfo() {
	doInUIThread(v.loadingInfoBar.start)
	v.account.session.LoadRoomInfo(v.roomID())
}

// roomConfigReceivedEvent MUST be called from the UI thread
func (v *roomView) roomConfigReceivedEvent(roomInfo data.RoomConfig) {
	v.loadingInfoBar.hide()

	v.warnings.clear()
	v.showRoomWarnings(roomInfo)
	v.notifications.add(v.warningsInfoBar)
}

// roomConfigRequestTimeoutEvent MUST be called from the UI thread
func (v *roomView) roomConfigRequestTimeoutEvent() {
	v.warnings.clear()

	v.loadingInfoBar.error(
		i18n.Local("An error occurred while loading room information"),
		i18n.Local("Loading the room information took longer than usual, perhaps the connection to the server was lost. Do you want to try again?."),
		v.requestRoomInfo,
	)
}

func (v *roomView) showRoomWarnings(roomInfo data.RoomConfig) {
	v.warnings.add(i18n.Local("Please be aware that communication in chat rooms is not encrypted - anyone that can intercept communication between you and the server - and the server itself - will be able to see what you are saying in this chat room. Only join this room and communicate here if you trust the server to not be hostile."))

	switch roomInfo.Anonymity {
	case "semi":
		v.warnings.add(i18n.Local("This room is partially anonymous. This means that only moderators can connect your nickname with your real username (your JID)."))
	case "no":
		v.warnings.add(i18n.Local("This room is not anonymous. This means that any person in this room can connect your nickname with your real username (your JID)."))
	default:
		v.log.WithField("anonymity", roomInfo.Anonymity).Warn("Unknown anonymity setting for room")
	}

	if roomInfo.Logged {
		v.warnings.add(i18n.Local("This room is publicly logged, meaning that everything you and the others in the room say or do can be made public on a website."))
	}
}

func (v *roomView) showWarnings() {
	prov := providerWithStyle("box", style{
		"background-color": "#ffffff",
		"box-shadow":       "0 10px 20px rgba(0, 0, 0, 0.35)",
	})

	updateWithStyle(v.messagesBox, prov)

	v.warnings.show()
	v.showNotificationsOverlay()
}

func (v *roomView) removeWarningsInfobar() {
	v.notifications.remove(v.warningsInfoBar.getWidget())
}

func (v *roomView) showNotificationsOverlay() {
	prov := providerWithStyle("box", style{
		"background-color": "rgba(0, 0, 0, 0.5)",
	})

	updateWithStyle(v.messagesOverlayBox, prov)

	v.messagesOverlay.Show()
}

func (v *roomView) closeNotificationsOverlay() {
	v.messagesOverlay.Hide()
}

func (v *roomView) onDestroyWindow() {
	v.opened = false
	v.account.removeRoomView(v.roomID())
	go v.cancelActiveRequests()
}

// cancelActiveRequests MUST NOT be called from the UI thread
func (v *roomView) cancelActiveRequests() {
	if v.cancel != nil {
		v.cancel <- true
		v.cancel = nil
	}
}

func (v *roomView) setTitle(t string) {
	v.window.SetTitle(t)
}

func (v *roomView) isOpen() bool {
	return v.opened
}

func (v *roomView) isJoined() bool {
	return v.room.SelfOccupantIsJoined()
}

func (v *roomView) isOwner() bool {
	return v.room.SelfOccupantIsOwner()
}

func (v *roomView) present() {
	if v.isOpen() {
		v.window.Present()
	}
}

func (v *roomView) show() {
	v.opened = true
	v.window.Show()
}

func (v *roomView) onLeaveRoom() {
	// TODO: Implement the logic behind leaving this room and
	// how the view will interact with the user during this process
	v.tryLeaveRoom(nil, nil)
}

func (v *roomView) tryLeaveRoom(onSuccess, onError func()) {
	v.spinner.show()

	go func() {
		v.account.leaveRoom(v.roomID(), v.room.SelfOccupantNickname(), func() {
			doInUIThread(v.window.Destroy)
			if onSuccess != nil {
				onSuccess()
			}
		}, func(err error) {
			v.log.WithError(err).Error("An error occurred when trying to leave the room")
			doInUIThread(v.spinner.hide)
			if onError != nil {
				onError()
			}
		})
	}()
}

func (v *roomView) onDestroyRoom() {
	d := v.newRoomDestroyView(v.window)
	d.show()
}

// tryDestroyRoom MUST be called from the UI thread, but please, note that
// the "onSuccess" and "onError" callbacks will be called from another thread
func (v *roomView) tryDestroyRoom(alternateID jid.Bare, reason string, onSuccess func(), onError func(error)) {
	v.spinner.show()

	onSuccessFinal := func() {
		doInUIThread(v.onRoomDestroyed)
		callFuncIfNotNil(onSuccess)
	}

	onErrorFinal := func(err error) {
		v.log.WithError(err).Error("An error occurred when trying to destroy the room")
		doInUIThread(v.spinner.hide)
		callFuncWithErrIfNotNil(onError, err)
	}

	go v.account.destroyRoom(v.roomID(), alternateID, reason, onSuccessFinal, onErrorFinal)
}

func (v *roomView) onRoomDestroyed() {
	v.notifications.info(i18n.Local("The room has been destroyed"))
}

func (v *roomView) switchToLobbyView() {
	v.initRoomLobby()

	if v.shouldReturnOnCancel() {
		v.lobby.switchToReturnOnCancel()
	} else {
		v.lobby.switchToCancel()
	}

	v.lobby.show()
}

func (v *roomView) switchToMainView() {
	v.initRoomMain()
	v.main.show()
}

func (v *roomView) onJoined() {
	doInUIThread(func() {
		v.lobby.hide()
		v.switchToMainView()
	})
}

func (v *roomView) shouldReturnOnCancel() bool {
	return v.returnTo != nil
}

func (v *roomView) onJoinCancel() {
	v.window.Destroy()

	if v.shouldReturnOnCancel() {
		v.returnTo()
	}
}

// messageForbidden MUST NOT be called from the UI thread
func (v *roomView) messageForbidden() {
	v.publishEvent(messageForbidden{})
}

// messageNotAccepted MUST NOT be called from the UI thread
func (v *roomView) messageNotAccepted() {
	v.publishEvent(messageNotAcceptable{})
}

// nicknameConflict MUST NOT be called from the UI thread
func (v *roomView) nicknameConflict(nickname string) {
	v.publishEvent(nicknameConflictEvent{nickname})
}

// registrationRequired MUST NOT be called from the UI thread
func (v *roomView) registrationRequired(nickname string) {
	v.publishEvent(registrationRequiredEvent{nickname})
}

func (v *roomView) roomID() jid.Bare {
	return v.room.ID
}

func (v *roomView) roomDisplayName() string {
	return v.roomID().Local().String()
}

func callFuncIfNotNil(f func()) {
	if f != nil {
		f()
	}
}

func callFuncWithErrIfNotNil(f func(error), err error) {
	if f != nil {
		f(err)
	}
}
