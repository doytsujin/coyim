package gtk_mock

import "github.com/twstrike/coyim/Godeps/_workspace/src/github.com/twstrike/gotk3adapter/gtki"

func init() {
	gtki.AssertGtk(&Mock{})
	gtki.AssertAboutDialog(&MockAboutDialog{})
	gtki.AssertAccelGroup(&MockAccelGroup{})
	gtki.AssertAdjustment(&MockAdjustment{})
	gtki.AssertApplication(&MockApplication{})
	gtki.AssertApplicationWindow(&MockApplicationWindow{})
	gtki.AssertBox(&MockBox{})
	gtki.AssertBuilder(&MockBuilder{})
	gtki.AssertButton(&MockButton{})
	gtki.AssertCellRenderer(&MockCellRenderer{})
	gtki.AssertCellRendererText(&MockCellRendererText{})
	gtki.AssertCellRendererToggle(&MockCellRendererToggle{})
	gtki.AssertCheckButton(&MockCheckButton{})
	gtki.AssertCheckMenuItem(&MockCheckMenuItem{})
	gtki.AssertComboBox(&MockComboBox{})
	gtki.AssertComboBoxText(&MockComboBoxText{})
	gtki.AssertCssProvider(&MockCssProvider{})
	gtki.AssertDialog(&MockDialog{})
	gtki.AssertEntry(&MockEntry{})
	gtki.AssertFileChooserDialog(&MockFileChooserDialog{})
	gtki.AssertGrid(&MockGrid{})
	gtki.AssertHeaderBar(&MockHeaderBar{})
	gtki.AssertInfoBar(&MockInfoBar{})
	gtki.AssertLabel(&MockLabel{})
	gtki.AssertListStore(&MockListStore{})
	gtki.AssertMenuBar(&MockMenuBar{})
	gtki.AssertMenuItem(&MockMenuItem{})
	gtki.AssertMenu(&MockMenu{})
	gtki.AssertMessageDialog(&MockMessageDialog{})
	gtki.AssertNotebook(&MockNotebook{})
	gtki.AssertRevealer(&MockRevealer{})
	gtki.AssertScrolledWindow(&MockScrolledWindow{})
	gtki.AssertSeparatorMenuItem(&MockSeparatorMenuItem{})
	gtki.AssertStyleContext(&MockStyleContext{})
	gtki.AssertTextBuffer(&MockTextBuffer{})
	gtki.AssertTextIter(&MockTextIter{})
	gtki.AssertTextTagTable(&MockTextTagTable{})
	gtki.AssertTextTag(&MockTextTag{})
	gtki.AssertTextView(&MockTextView{})
	gtki.AssertTreeIter(&MockTreeIter{})
	gtki.AssertTreePath(&MockTreePath{})
	gtki.AssertTreeSelection(&MockTreeSelection{})
	gtki.AssertTreeStore(&MockTreeStore{})
	gtki.AssertTreeView(&MockTreeView{})
	gtki.AssertTreeViewColumn(&MockTreeViewColumn{})
	gtki.AssertWidget(&MockWidget{})
	gtki.AssertWindow(&MockWindow{})
}
