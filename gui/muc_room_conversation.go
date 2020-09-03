package gui

import (
	"fmt"

	"github.com/coyim/coyim/i18n"
	"github.com/coyim/coyim/xmpp/jid"
	"github.com/coyim/gotk3adapter/gtki"
	"github.com/coyim/gotk3adapter/pangoi"
)

type roomViewConversation struct {
	view             gtki.Box      `gtk-widget:"roomConversation"`
	roomChatTextView gtki.TextView `gtk-widget:"roomChatTextView"`

	tags *mucStyleTags
}

type mucStyleTags struct {
	table gtki.TextTagTable
	buf   gtki.TextBuffer
}

func (v *roomViewConversation) getStyleTags() *mucStyleTags {
	if v.tags == nil {
		v.tags = v.newStyleTags()
	}
	return v.tags
}

func (v *roomViewConversation) newStyleTags() *mucStyleTags {
	// TODO: for now we are using a default styles, but we can improve it
	// if we define a structure with a predefined colors pallete based on kind
	// of messages to show like entering a room, leaving the room, incoming
	// message, etc
	t := new(mucStyleTags)

	t.table, _ = g.gtk.TextTagTableNew()

	leftRoomTag, _ := g.gtk.TextTagNew("leftRoomText")
	_ = leftRoomTag.SetProperty("foreground", "#EE0000")
	_ = leftRoomTag.SetProperty("style", pangoi.STYLE_ITALIC)

	t.table.Add(leftRoomTag)

	t.buf, _ = g.gtk.TextBufferNew(t.table)

	return t
}

func newRoomViewConversation() *roomViewConversation {
	c := &roomViewConversation{}

	builder := newBuilder("MUCRoomConversation")
	panicOnDevError(builder.bindObjects(c))

	t := c.getStyleTags()
	c.roomChatTextView.SetBuffer(t.buf)

	return c
}

func (v *roomViewConversation) showOccupantLeftRoom(nickname jid.Resource) {
	doInUIThread(func() {
		msg := i18n.Localf("%s left the room", nickname)
		v.addLineToChatTextUsingTagID(msg, "leftRoomText")
	})
}

func (v *roomViewConversation) addLineToChatText(text string) {
	buf, _ := v.roomChatTextView.GetBuffer()
	i := buf.GetEndIter()

	buf.Insert(i, fmt.Sprintf("%s\n", text))
}

func (v *roomViewConversation) addLineToChatTextUsingTagID(text string, tag string) {
	buf, _ := v.roomChatTextView.GetBuffer()

	charCount := buf.GetCharCount()
	v.addLineToChatText(text)
	oldIterEnd := buf.GetIterAtOffset(charCount)
	newIterEnd := buf.GetEndIter()

	buf.ApplyTagByName(tag, oldIterEnd, newIterEnd)
}
