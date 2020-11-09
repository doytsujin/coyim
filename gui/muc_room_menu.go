package gui

import (
	"fmt"

	"github.com/coyim/coyim/i18n"
	"github.com/coyim/gotk3adapter/gtki"
)

type roomViewMenuItem interface {
	getRoomViewMenuItemWidget() gtki.Widget
}

type roomViewMenuButton struct {
	label   string
	onClick func()

	button gtki.ModelButton `gtk-widget:"room-menu-button"`
}

func newRoomViewMenuButton(l string, onClick func()) *roomViewMenuButton {
	mb := &roomViewMenuButton{
		label:   l,
		onClick: wrapSafeCallOfNilFunc(onClick),
	}

	b := newBuilder("MUCRoomMenuButton")
	panicOnDevError(b.bindObjects(mb))

	b.ConnectSignals(map[string]interface{}{
		"on_clicked": mb.onClick,
	})

	mb.button.SetLabel(mb.label)

	return mb
}

// implements roomViewMenuItem interface
func (b *roomViewMenuButton) getRoomViewMenuItemWidget() gtki.Widget {
	return b.button
}

type roomViewMenuDivider struct {
	s gtki.Separator
}

// implements roomViewMenuItem interface
func (d *roomViewMenuDivider) getRoomViewMenuItemWidget() gtki.Widget {
	return d.s
}

func newRoomViewMenuDivider() *roomViewMenuDivider {
	d, _ := g.gtk.SeparatorNew(gtki.HorizontalOrientation)
	return &roomViewMenuDivider{d}
}

type roomViewMenu struct {
	items map[string]roomViewMenuItem

	popover gtki.Popover `gtk-widget:"room-menu"`
	view    gtki.Box     `gtk-widget:"room-menu-box"`
}

// newRoomViewMenu MUST be called from the UI thread
func newRoomViewMenu() *roomViewMenu {
	m := &roomViewMenu{
		items: make(map[string]roomViewMenuItem),
	}

	m.initBuilder()

	return m
}

func (m *roomViewMenu) initBuilder() {
	b := newBuilder("MUCRoomMenu")
	panicOnDevError(b.bindObjects(m))
}

// addOrUpdateMenuItem MUST always be called from the UI thread
func (m *roomViewMenu) addOrUpdateMenuItem(id string, item roomViewMenuItem) {
	m.items[id] = item
	m.redraw()
}

// addButtonItem MUST always be called from the UI thread
func (m *roomViewMenu) addButtonItem(id, l string, f func()) {
	m.addOrUpdateMenuItem(id, newRoomViewMenuButton(l, f))
}

// addDividerItem MUST always be called from the UI thread
func (m *roomViewMenu) addDividerItem() {
	m.addOrUpdateMenuItem(fmt.Sprintf("divider-%d", len(m.items)+1), newRoomViewMenuDivider())
}

// redraw MUST be called from the UI thread
func (m *roomViewMenu) redraw() {
	m.removeAll()

	for _, i := range m.items {
		m.view.Add(i.getRoomViewMenuItemWidget())
	}
}

// removeAll MUST be called from the UI thread
func (m *roomViewMenu) removeAll() {
	for _, i := range m.items {
		m.view.Remove(i.getRoomViewMenuItemWidget())
	}
}

// reset MUST be called from the UI thread
//
// The difference between this method and "removeAll" is
// that this method will remove all items from the list and from the view,
// while "removeAll" only will remove elements from the view
func (m *roomViewMenu) reset() {
	m.removeAll()
	m.items = make(map[string]roomViewMenuItem)
}

// initRoomMenu MUST be called from the UI thread
func (v *roomView) initRoomMenu() {
	v.menu = newRoomViewMenu()
	v.refreshRoomMenu()
}

// refreshRoomMenu MUST be called from the UI thread
func (v *roomView) refreshRoomMenu() {
	v.menu.reset()

	if v.isSelfOccupantAnOwner() {
		v.menu.addButtonItem("destroy-room", i18n.Local("Destroy room"), v.onDestroyRoom)
		v.menu.addDividerItem()
	}

	if v.isSelfOccupantJoined() {
		v.menu.addButtonItem("leave-room", i18n.Local("Leave room"), v.onLeaveRoom)
	}
}

func (v *roomView) getRoomMenuWidget() gtki.Popover {
	return v.menu.popover
}