package gui

import (
	"sync"

	"github.com/coyim/coyim/coylog"
	"github.com/coyim/coyim/session/events"
	"github.com/coyim/coyim/session/muc"
	"github.com/coyim/coyim/xmpp/jid"
	log "github.com/sirupsen/logrus"
)

func (v *roomView) handleRoomEvent(ev muc.MUC) {
	switch t := ev.(type) {
	case events.MUCSelfOccupantJoined:
		doInUIThread(func() {
			v.publishWithInfo(occupantSelfJoined, roomViewEventInfo{
				nickname: t.Nickname,
			})
		})

	case events.MUCOccupantUpdated:
		doInUIThread(func() {
			v.publishWithInfo(occupantUpdated, roomViewEventInfo{
				nickname: t.Nickname,
			})
		})

	case events.MUCOccupantJoined:
		doInUIThread(func() {
			v.publishWithInfo(occupantJoined, roomViewEventInfo{
				nickname: t.Nickname,
			})
		})

	case events.MUCOccupantLeft:
		doInUIThread(func() {
			v.publishWithInfo(occupantLeft, roomViewEventInfo{
				nickname: t.Nickname,
			})
		})

	case events.MUCMessageReceived:
		doInUIThread(func() {
			v.publishWithInfo(messageReceived, roomViewEventInfo{
				nickname: t.Nickname,
				subject:  t.Subject,
				message:  t.Message,
			})
		})

	case events.MUCLoggingEnabled:
		doInUIThread(func() {
			v.publish(loggingEnabled)
		})

	case events.MUCLoggingDisabled:
		doInUIThread(func() {
			v.publish(loggingDisabled)
		})

	default:
		v.log.WithField("event", t).Warn("Unsupported room event received")
	}
}

func (v *roomView) observeRoomEvents() {
	for ev := range v.events {
		v.handleRoomEvent(ev)
	}
}

// roomViewEventInfo contains information about any room view event
type roomViewEventInfo struct {
	nickname string
	subject  string
	message  string
}

type roomViewEventType int

const (
	occupantSelfJoined roomViewEventType = iota
	occupantJoined
	occupantUpdated
	occupantLeft

	roomInfoReceived

	messageReceived

	loggingEnabled
	loggingDisabled

	registrationRequired
	nicknameConflict

	previousToSwitchToLobby
	previousToSwitchToMain
)

type roomViewObserver struct {
	id       string
	onNotify func(roomViewEventInfo)
}

type roomViewSubscribers struct {
	sync.RWMutex
	log       coylog.Logger
	observers map[roomViewEventType][]*roomViewObserver
}

func newRoomViewSubscribers(room jid.Bare, l coylog.Logger) *roomViewSubscribers {
	return &roomViewSubscribers{
		log: l.WithFields(log.Fields{
			"who":  "newRoomViewSubscribers",
			"room": room,
		}),
		observers: make(map[roomViewEventType][]*roomViewObserver),
	}
}

func (s *roomViewSubscribers) subscribe(ev roomViewEventType, o *roomViewObserver) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.observers[ev]; !ok {
		s.observers[ev] = make([]*roomViewObserver, 0)
	}

	s.observers[ev] = append(s.observers[ev], o)
}

func (s *roomViewSubscribers) unsubscribe(ev roomViewEventType, id string) {
	s.RLock()
	defer s.RUnlock()

	observers, ok := s.observers[ev]
	if !ok {
		s.debug("unsubscribe(): trying to unsubscribe from a not initialized event", ev)
		return
	}

	for i, o := range observers {
		if o.id == id {
			s.observers[ev] = append(
				s.observers[ev][:i], s.observers[ev][i+1:]...,
			)
			return
		}
	}
}

func (s *roomViewSubscribers) publish(ev roomViewEventType, ei roomViewEventInfo) {
	s.RLock()
	defer s.RUnlock()

	observers, ok := s.observers[ev]
	if !ok {
		s.debug("publish(): trying to publish a not initialized event", ev)
		return
	}

	for _, o := range observers {
		o.onNotify(ei)
	}
}

func (s *roomViewSubscribers) debug(m string, ev roomViewEventType) {
	s.log.Debug(m, ev)
}

func (v *roomView) subscribe(id string, ev roomViewEventType, onNotify func(roomViewEventInfo)) {
	v.subscribers.subscribe(ev, &roomViewObserver{
		id:       id,
		onNotify: onNotify,
	})
}

func (v *roomView) unsubscribe(id string, ev roomViewEventType) {
	v.subscribers.unsubscribe(ev, id)
}

func (v *roomView) publish(ev roomViewEventType) {
	v.subscribers.publish(ev, roomViewEventInfo{})
}

func (v *roomView) publishWithInfo(ev roomViewEventType, ei roomViewEventInfo) {
	v.subscribers.publish(ev, ei)
}
