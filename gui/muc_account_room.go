package gui

import "github.com/coyim/coyim/xmpp/jid"

// leaveRoom should return the context so the caller can cancel the context early if required
func (a *account) leaveRoom(roomID jid.Bare, nickname string, onSuccess func(), onError func(error)) {
	leaveRoom := func() (<-chan bool, <-chan error, func()) {
		ok, err := a.session.LeaveRoom(roomID, nickname)
		return ok, err, nil
	}

	leaveRoomSuccess := func() {
		a.removeRoomView(roomID)
		if onSuccess != nil {
			onSuccess()
		}
	}

	controller := a.newRoomOpController("leave-room", leaveRoom, leaveRoomSuccess, onError)
	ctx := a.newAccountRoomOpContext("leave-room", roomID, controller)

	go ctx.doOperation()
}

// destroyRoom should return the context so the caller can cancel the context early if required
func (a *account) destroyRoom(roomID jid.Bare, alternativeRoomID jid.Bare, reason string, onSuccess func(), onError func(error)) {
	destroyRoom := func() (<-chan bool, <-chan error, func()) {
		return a.session.DestroyRoom(roomID, alternativeRoomID, reason)
	}

	controller := a.newRoomOpController("destroy-room", destroyRoom, onSuccess, onError)
	ctx := a.newAccountRoomOpContext("destroy-room", roomID, controller)

	go ctx.doOperation()
}