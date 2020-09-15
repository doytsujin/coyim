package gui

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/coyim/gotk3adapter/gtk_mock"
	"github.com/coyim/gotk3adapter/gtki"

	. "gopkg.in/check.v1"
)

type UIReaderSuite struct{}

var _ = Suite(&UIReaderSuite{})

const testFile string = `
<interface>
  <object class="GtkWindow" id="conversation">
    <property name="default-height">600</property>
    <property name="default-width">500</property>
    <child>
	  <object class="GtkBox" id="vbox">
	    <property name="orientation">GTK_ORIENTATION_VERTICAL</property>  
	  </object>
    </child>
  </object>
</interface>
`

func writeTestFile(name, content string) {
	_ = ioutil.WriteFile(name, []byte(content), 0700)
}

func removeFile(name string) {
	_ = os.Remove(name)
}

type mockBuilder struct {
	gtk_mock.MockBuilder
	stringGiven string
	errorGiven  error
}

func (v *mockBuilder) AddFromString(v1 string) error {
	v.stringGiven = v1
	return v.errorGiven
}

type mockWithBuilder struct {
	gtk_mock.Mock
	errorGiven       error
	secondErrorGiven error
}

func (v *mockWithBuilder) BuilderNew() (gtki.Builder, error) {
	return &mockBuilder{
		errorGiven: v.secondErrorGiven,
	}, v.errorGiven
}

const wrongTemplate string = `
yeah
<interfae>
  <object class="GtkWindow">
    I have a bad format
  </object>
`

func (s *UIReaderSuite) Test_builderForString_panicsIfNotBuilder(c *C) {
	g = Graphics{gtk: &mockWithBuilder{
		errorGiven: errors.New("bla"),
	}}

	c.Assert(func() {
		_ = builderForString(testFile)
	}, PanicMatches, "bla")
}

func (s *UIReaderSuite) Test_builderForString_panicsIfEmptyTemplate(c *C) {
	g = Graphics{gtk: &mockWithBuilder{
		secondErrorGiven: errors.New("foo"),
	}}

	c.Assert(func() {
		builderForString("")
	}, PanicMatches, "gui: wrong template format: foo\n")
}

func (s *UIReaderSuite) Test_builderForString_panicsIfWrongTemplate(c *C) {
	g = Graphics{gtk: &mockWithBuilder{
		secondErrorGiven: errors.New("bla"),
	}}

	c.Assert(func() {
		builderForString(wrongTemplate)
	}, PanicMatches, "gui: wrong template format: bla\n")
}

func (s *UIReaderSuite) Test_builderForString_useTemplateStringIfOk(c *C) {
	g = Graphics{gtk: &mockWithBuilder{}}

	builder := builderForString(testFile)

	str := builder.(*mockBuilder).stringGiven

	c.Assert(str, Equals, testFile)
}

func (s *UIReaderSuite) Test_builderForDefinition_useXMLIfExists(c *C) {
	orgMod := getModTime(getActualDefsFolder() + "/Test.xml")
	defer func() {
		setModTime(getActualDefsFolder()+"/Test.xml", orgMod)
	}()

	g = Graphics{gtk: &mockWithBuilder{}}
	removeFile(getActualDefsFolder() + "/Test.xml")
	writeTestFile(getActualDefsFolder()+"/Test.xml", testFile)
	ui := "Test"

	builder := builderForDefinition(ui)

	str := builder.(*mockBuilder).stringGiven

	c.Assert(str, Equals, testFile)
}

func getModTime(fn string) time.Time {
	file, _ := os.Stat(fn)
	return file.ModTime()
}

func setModTime(fn string, t time.Time) {
	_ = os.Chtimes(fn, t, t)
}

func (s *UIReaderSuite) Test_builderForDefinition_useGoFileIfXMLDoesntExists(c *C) {
	orgMod := getModTime(getActualDefsFolder() + "/Test.xml")
	defer func() {
		writeTestFile(getActualDefsFolder()+"/Test.xml", testFile)
		setModTime(getActualDefsFolder()+"/Test.xml", orgMod)
	}()

	g = Graphics{gtk: &mockWithBuilder{}}
	removeFile(getActualDefsFolder() + "/Test.xml")
	ui := "Test"

	builder := builderForDefinition(ui)

	str := builder.(*mockBuilder).stringGiven

	c.Assert(str, Equals, testFile)
}

func (s *UIReaderSuite) Test_builderForDefinition_shouldReturnErrorWhenDefinitionDoesntExist(c *C) {
	ui := "nonexistent"

	c.Assert(func() {
		builderForDefinition(ui)
	}, Panics, "No definition found for nonexistent")
}

func (s *UIReaderSuite) Test_getImageBytes_forExistingImage(c *C) {
	r := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(
		`
PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiIHN0YW5kYWxvbmU9Im5vIj8+Cjxzdmcg
d2lkdGg9IjE1cHgiIGhlaWdodD0iMTRweCIgdmlld0JveD0iMCAwIDE1IDE0IiB2ZXJzaW9uPSIxLjEi
IHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cu
dzMub3JnLzE5OTkveGxpbmsiPgogICAgPCEtLSBHZW5lcmF0b3I6IFNrZXRjaCA0MS4yICgzNTM5Nykg
LSBodHRwOi8vd3d3LmJvaGVtaWFuY29kaW5nLmNvbS9za2V0Y2ggLS0+CiAgICA8dGl0bGU+74GxIGNv
cHkgMjwvdGl0bGU+CiAgICA8ZGVzYz5DcmVhdGVkIHdpdGggU2tldGNoLjwvZGVzYz4KICAgIDxkZWZz
PjwvZGVmcz4KICAgIDxnIGlkPSJTTVAtZmxvd3MiIHN0cm9rZT0ibm9uZSIgc3Ryb2tlLXdpZHRoPSIx
IiBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPgogICAgICAgIDxnIGlkPSJCb2JfNGIiIHRy
YW5zZm9ybT0idHJhbnNsYXRlKC04MzkuMDAwMDAwLCAtMTY0LjAwMDAwMCkiIGZpbGw9IiNGODlCMUMi
PgogICAgICAgICAgICA8ZyBpZD0iR3JvdXAtMi1Db3B5IiB0cmFuc2Zvcm09InRyYW5zbGF0ZSg2MDAu
MDAwMDAwLCAxNDAuMDAwMDAwKSI+CiAgICAgICAgICAgICAgICA8cGF0aCBkPSJNMjQ3LjU3MDIxNywz
NS40OTY0NzY0IEwyNDcuNTcwMjE3LDMzLjkwNzg3MjQgQzI0Ny41NzAyMTcsMzMuODI5ODM1MyAyNDcu
NTQzNzQxLDMzLjc2NDM0MDkgMjQ3LjQ5MDc4NywzMy43MTEzODcxIEMyNDcuNDM3ODM0LDMzLjY1ODQz
MzQgMjQ3LjM3NTEyNiwzMy42MzE5NTY5IDI0Ny4zMDI2NjMsMzMuNjMxOTU2OSBMMjQ1LjY5NzMzNywz
My42MzE5NTY5IEMyNDUuNjI0ODc0LDMzLjYzMTk1NjkgMjQ1LjU2MjE2NiwzMy42NTg0MzM0IDI0NS41
MDkyMTMsMzMuNzExMzg3MSBDMjQ1LjQ1NjI1OSwzMy43NjQzNDA5IDI0NS40Mjk3ODMsMzMuODI5ODM1
MyAyNDUuNDI5NzgzLDMzLjkwNzg3MjQgTDI0NS40Mjk3ODMsMzUuNDk2NDc2NCBDMjQ1LjQyOTc4Mywz
NS41NzQ1MTM0IDI0NS40NTYyNTksMzUuNjQwMDA3OSAyNDUuNTA5MjEzLDM1LjY5Mjk2MTYgQzI0NS41
NjIxNjYsMzUuNzQ1OTE1MyAyNDUuNjI0ODc0LDM1Ljc3MjM5MTggMjQ1LjY5NzMzNywzNS43NzIzOTE4
IEwyNDcuMzAyNjYzLDM1Ljc3MjM5MTggQzI0Ny4zNzUxMjYsMzUuNzcyMzkxOCAyNDcuNDM3ODM0LDM1
Ljc0NTkxNTMgMjQ3LjQ5MDc4NywzNS42OTI5NjE2IEMyNDcuNTQzNzQxLDM1LjY0MDAwNzkgMjQ3LjU3
MDIxNywzNS41NzQ1MTM0IDI0Ny41NzAyMTcsMzUuNDk2NDc2NCBMMjQ3LjU3MDIxNywzNS40OTY0NzY0
IFogTTI0Ny41NTM0OTUsMzIuMzY5NDM0OCBMMjQ3LjcwMzk5NSwyOC41MzE3MDIgQzI0Ny43MDM5OTUs
MjguNDY0ODEzIDI0Ny42NzYxMjUsMjguNDExODYwMSAyNDcuNjIwMzg0LDI4LjM3Mjg0MTYgQzI0Ny41
NDc5MjEsMjguMzExNTI2NyAyNDcuNDgxMDMzLDI4LjI4MDg2OTcgMjQ3LjQxOTcxOCwyOC4yODA4Njk3
IEwyNDUuNTgwMjgyLDI4LjI4MDg2OTcgQzI0NS41MTg5NjcsMjguMjgwODY5NyAyNDUuNDUyMDc5LDI4
LjMxMTUyNjcgMjQ1LjM3OTYxNiwyOC4zNzI4NDE2IEMyNDUuMzIzODc1LDI4LjQxMTg2MDEgMjQ1LjI5
NjAwNSwyOC40NzAzODcgMjQ1LjI5NjAwNSwyOC41NDg0MjQxIEwyNDUuNDM4MTQ0LDMyLjM2OTQzNDgg
QzI0NS40MzgxNDQsMzIuNDI1MTc1NiAyNDUuNDY2MDE0LDMyLjQ3MTE2MSAyNDUuNTIxNzU0LDMyLjUw
NzM5MjUgQzI0NS41Nzc0OTUsMzIuNTQzNjI0IDI0NS42NDQzODMsMzIuNTYxNzM5NSAyNDUuNzIyNDIs
MzIuNTYxNzM5NSBMMjQ3LjI2OTIxOSwzMi41NjE3Mzk1IEMyNDcuMzQ3MjU2LDMyLjU2MTczOTUgMjQ3
LjQxMjc1LDMyLjU0MzYyNCAyNDcuNDY1NzA0LDMyLjUwNzM5MjUgQzI0Ny41MTg2NTgsMzIuNDcxMTYx
IDI0Ny41NDc5MjEsMzIuNDI1MTc1NiAyNDcuNTUzNDk1LDMyLjM2OTQzNDggTDI0Ny41NTM0OTUsMzIu
MzY5NDM0OCBaIE0yNDcuNDM2NDQsMjQuNTYwMTkxOSBMMjUzLjg1Nzc0NSwzNi4zMzI1ODM3IEMyNTQu
MDUyODM4LDM2LjY4Mzc1MDYgMjU0LjA0NzI2NCwzNy4wMzQ5MTIyIDI1My44NDEwMjMsMzcuMzg2MDc5
IEMyNTMuNzQ2MjYzLDM3LjU0NzcyNzMgMjUzLjYxNjY2OCwzNy42NzU5MjkxIDI1My40NTIyMzMsMzcu
NzcwNjg4NCBDMjUzLjI4Nzc5OCwzNy44NjU0NDc3IDI1My4xMTA4MjMsMzcuOTEyODI2NyAyNTIuOTIx
MzA1LDM3LjkxMjgyNjcgTDI0MC4wNzg2OTUsMzcuOTEyODI2NyBDMjM5Ljg4OTE3NywzNy45MTI4MjY3
IDIzOS43MTIyMDIsMzcuODY1NDQ3NyAyMzkuNTQ3NzY3LDM3Ljc3MDY4ODQgQzIzOS4zODMzMzIsMzcu
Njc1OTI5MSAyMzkuMjUzNzM3LDM3LjU0NzcyNzMgMjM5LjE1ODk3NywzNy4zODYwNzkgQzIzOC45NTI3
MzYsMzcuMDM0OTEyMiAyMzguOTQ3MTYyLDM2LjY4Mzc1MDYgMjM5LjE0MjI1NSwzNi4zMzI1ODM3IEwy
NDUuNTYzNTYsMjQuNTYwMTkxOSBDMjQ1LjY1ODMxOSwyNC4zODczOTU2IDI0NS43ODkzMDgsMjQuMjUw
ODMyNyAyNDUuOTU2NTMsMjQuMTUwNDk5MyBDMjQ2LjEyMzc1MywyNC4wNTAxNjU5IDI0Ni4zMDQ5MDcs
MjQgMjQ2LjUsMjQgQzI0Ni42OTUwOTMsMjQgMjQ2Ljg3NjI0NywyNC4wNTAxNjU5IDI0Ny4wNDM0Nywy
NC4xNTA0OTkzIEMyNDcuMjEwNjkyLDI0LjI1MDgzMjcgMjQ3LjM0MTY4MSwyNC4zODczOTU2IDI0Ny40
MzY0NCwyNC41NjAxOTE5IEwyNDcuNDM2NDQsMjQuNTYwMTkxOSBaIiBpZD0i74GxLWNvcHktMiI+PC9w
YXRoPgogICAgICAgICAgICA8L2c+CiAgICAgICAgPC9nPgogICAgPC9nPgo8L3N2Zz4=`))
	expectedBytes, _ := ioutil.ReadAll(r)

	c.Assert(expectedBytes, DeepEquals, mustGetImageBytes("alert.svg"))

	r = base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(
		`
PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiIHN0YW5kYWxvbmU9Im5vIj8+Cjxzdmcg
d2lkdGg9IjExcHgiIGhlaWdodD0iMTNweCIgdmlld0JveD0iMCAwIDExIDEzIiB2ZXJzaW9uPSIxLjEi
IHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cu
dzMub3JnLzE5OTkveGxpbmsiPgogICAgPCEtLSBHZW5lcmF0b3I6IFNrZXRjaCA0MS4yICgzNTM5Nykg
LSBodHRwOi8vd3d3LmJvaGVtaWFuY29kaW5nLmNvbS9za2V0Y2ggLS0+CiAgICA8dGl0bGU+R3JvdXA8
L3RpdGxlPgogICAgPGRlc2M+Q3JlYXRlZCB3aXRoIFNrZXRjaC48L2Rlc2M+CiAgICA8ZGVmcz48L2Rl
ZnM+CiAgICA8ZyBpZD0iU01QLWZsb3dzIiBzdHJva2U9Im5vbmUiIHN0cm9rZS13aWR0aD0iMSIgZmls
bD0ibm9uZSIgZmlsbC1ydWxlPSJldmVub2RkIj4KICAgICAgICA8ZyBpZD0iQm9iXzQiIHRyYW5zZm9y
bT0idHJhbnNsYXRlKC04NDMuMDAwMDAwLCAtMTY0LjAwMDAwMCkiPgogICAgICAgICAgICA8ZyBpZD0i
R3JvdXAiIHRyYW5zZm9ybT0idHJhbnNsYXRlKDg0My4wMDAwMDAsIDE2NC4wMDAwMDApIj4KICAgICAg
ICAgICAgICAgIDxwYXRoIGQ9Ik0zLjA1NTU1NTU2LDUuOTA5MDkwOTEgTDcuOTQ0NDQ0NDQsNS45MDkw
OTA5MSBMNy45NDQ0NDQ0NCw0LjEzNjM2MzY0IEM3Ljk0NDQ0NDQ0LDMuNDgzODk4MjUgNy43MDU3MzE1
NSwyLjkyNjg0ODkgNy4yMjgyOTg2MSwyLjQ2NTE5ODg2IEM2Ljc1MDg2NTY3LDIuMDAzNTQ4ODMgNi4x
NzQ3NzE4OSwxLjc3MjcyNzI3IDUuNSwxLjc3MjcyNzI3IEM0LjgyNTIyODExLDEuNzcyNzI3MjcgNC4y
NDkxMzQzMywyLjAwMzU0ODgzIDMuNzcxNzAxMzksMi40NjUxOTg4NiBDMy4yOTQyNjg0NSwyLjkyNjg0
ODkgMy4wNTU1NTU1NiwzLjQ4Mzg5ODI1IDMuMDU1NTU1NTYsNC4xMzYzNjM2NCBMMy4wNTU1NTU1Niw1
LjkwOTA5MDkxIFogTTExLDYuNzk1NDU0NTUgTDExLDEyLjExMzYzNjQgQzExLDEyLjM1OTg0OTcgMTAu
OTEwODgwNSwxMi41NjkxMjc5IDEwLjczMjYzODksMTIuNzQxNDc3MyBDMTAuNTU0Mzk3MywxMi45MTM4
MjY2IDEwLjMzNzk2NDIsMTMgMTAuMDgzMzMzMywxMyBMMC45MTY2NjY2NjcsMTMgQzAuNjYyMDM1NzY0
LDEzIDAuNDQ1NjAyNzQzLDEyLjkxMzgyNjYgMC4yNjczNjExMTEsMTIuNzQxNDc3MyBDMC4wODkxMTk0
NzkyLDEyLjU2OTEyNzkgMCwxMi4zNTk4NDk3IDAsMTIuMTEzNjM2NCBMMCw2Ljc5NTQ1NDU1IEMwLDYu
NTQ5MjQxMTkgMC4wODkxMTk0NzkyLDYuMzM5OTYyOTggMC4yNjczNjExMTEsNi4xNjc2MTM2NCBDMC40
NDU2MDI3NDMsNS45OTUyNjQyOSAwLjY2MjAzNTc2NCw1LjkwOTA5MDkxIDAuOTE2NjY2NjY3LDUuOTA5
MDkwOTEgTDEuMjIyMjIyMjIsNS45MDkwOTA5MSBMMS4yMjIyMjIyMiw0LjEzNjM2MzY0IEMxLjIyMjIy
MjIyLDMuMDAzNzgyMjIgMS42NDIzNTY5MSwyLjAzMTI1NDA2IDIuNDgyNjM4ODksMS4yMTg3NSBDMy4z
MjI5MjA4NywwLjQwNjI0NTkzNyA0LjMyODY5Nzg1LDAgNS41LDAgQzYuNjcxMzAyMTUsMCA3LjY3NzA3
OTEzLDAuNDA2MjQ1OTM3IDguNTE3MzYxMTEsMS4yMTg3NSBDOS4zNTc2NDMwOSwyLjAzMTI1NDA2IDku
Nzc3Nzc3NzgsMy4wMDM3ODIyMiA5Ljc3Nzc3Nzc4LDQuMTM2MzYzNjQgTDkuNzc3Nzc3NzgsNS45MDkw
OTA5MSBMMTAuMDgzMzMzMyw1LjkwOTA5MDkxIEMxMC4zMzc5NjQyLDUuOTA5MDkwOTEgMTAuNTU0Mzk3
Myw1Ljk5NTI2NDI5IDEwLjczMjYzODksNi4xNjc2MTM2NCBDMTAuOTEwODgwNSw2LjMzOTk2Mjk4IDEx
LDYuNTQ5MjQxMTkgMTEsNi43OTU0NTQ1NSBMMTEsNi43OTU0NTQ1NSBaIiBpZD0i74CjLWNvcHktMiIg
ZmlsbD0iIzdFRDMyMSI+PC9wYXRoPgogICAgICAgICAgICAgICAgPHBhdGggZD0iTTguMzIzMjMxNzEs
OC4wMDQ5NDQ4NyBMNS4yMjA1NjI4MywxMS4wMDUwNDYyIEM1LjE1OTU3NjQxLDExLjA2NDAxNjUgNS4w
ODcxNTYxMSwxMS4wOTM1MDEzIDUuMDAzMjk5NzgsMTEuMDkzNTAxMyBDNC45MTk0NDM0NCwxMS4wOTM1
MDEzIDQuODQ3MDIzMTUsMTEuMDY0MDE2NSA0Ljc4NjAzNjcyLDExLjAwNTA0NjIgTDMuMTQ3MDM0NzQs
OS40MjAyMjYwOCBDMy4wODYwNDgzMSw5LjM2MTI1NTczIDMuMDU1NTU1NTYsOS4yOTEyMjk1IDMuMDU1
NTU1NTYsOS4yMTAxNDUyNyBDMy4wNTU1NTU1Niw5LjEyOTA2MTA1IDMuMDg2MDQ4MzEsOS4wNTkwMzQ4
MiAzLjE0NzAzNDc0LDkuMDAwMDY0NDcgTDMuNTY2MzE0MzEsOC41OTQ2NDUzNyBDMy42MjczMDA3NCw4
LjUzNTY3NTAzIDMuNjk5NzIxMDMsOC41MDYxOTAzIDMuNzgzNTc3MzcsOC41MDYxOTAzIEMzLjg2NzQz
MzcsOC41MDYxOTAzIDMuOTM5ODU0LDguNTM1Njc1MDMgNC4wMDA4NDA0Miw4LjU5NDY0NTM3IEw1LjAw
MzI5OTc4LDkuNTYzOTY1NTggTDcuNDY5NDI2MDIsNy4xNzkzNjQxNyBDNy41MzA0MTI0NSw3LjEyMDM5
MzgyIDcuNjAyODMyNzQsNy4wOTA5MDkwOSA3LjY4NjY4OTA4LDcuMDkwOTA5MDkgQzcuNzcwNTQ1NDEs
Ny4wOTA5MDkwOSA3Ljg0Mjk2NTcsNy4xMjAzOTM4MiA3LjkwMzk1MjEzLDcuMTc5MzY0MTcgTDguMzIz
MjMxNzEsNy41ODQ3ODMyNiBDOC4zODQyMTgxMyw3LjY0Mzc1MzYxIDguNDE0NzEwODksNy43MTM3Nzk4
NCA4LjQxNDcxMDg5LDcuNzk0ODY0MDcgQzguNDE0NzEwODksNy44NzU5NDgyOSA4LjM4NDIxODEzLDcu
OTQ1OTc0NTMgOC4zMjMyMzE3MSw4LjAwNDk0NDg3IEw4LjMyMzIzMTcxLDguMDA0OTQ0ODcgWiIgaWQ9
IlBhdGgiIGZpbGw9IiNGRkZGRkYiPjwvcGF0aD4KICAgICAgICAgICAgPC9nPgogICAgICAgIDwvZz4K
ICAgIDwvZz4KPC9zdmc+`))
	expectedBytes, _ = ioutil.ReadAll(r)

	c.Assert(expectedBytes, DeepEquals, mustGetImageBytes("padlock.svg"))
}

func (s *UIReaderSuite) Test_GettingNonExistantImage_Panics(c *C) {
	image := "nonexistent"

	c.Assert(func() {
		mustGetImageBytes(image)
	}, Panics, "Developer error: getting the image "+image+" but it does not exist")
}
