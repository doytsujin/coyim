package gui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twstrike/coyim/gui/definitions"
)

const (
	defsFolder   = "gui/definitions"
	xmlExtension = ".xml"
)

func getDefinitionWithFileFallback(uiName string) string {
	// this makes sure a missing definition wont break only when the app is released
	uiDef := getDefinition(uiName)

	fileName := filepath.Join(defsFolder, uiName+xmlExtension)
	if fileNotFound(fileName) {
		log.Printf("gui: loading compiled definition %q\n", uiName)
		return uiDef.String()
	}

	return readFile(fileName)
}

func loadBuilderWith(uiName string) (*gtk.Builder, error) {
	template := getDefinitionWithFileFallback(uiName)

	builder, err := gtk.BuilderNew()
	if err != nil {
		return nil, err
	}

	err = builder.AddFromString(template)
	if err != nil {
		log.Printf("gui: failed load %s: %s\n", uiName, err.Error())
		return nil, err
	}

	return builder, nil
}

func fileNotFound(fileName string) bool {
	_, fnf := os.Stat(fileName)
	return os.IsNotExist(fnf)
}

func readFile(fileName string) string {
	file, _ := os.Open(fileName)
	reader := bufio.NewScanner(file)
	var content string
	for reader.Scan() {
		content = content + reader.Text()
	}
	file.Close()
	return content
}

func getDefinition(uiName string) fmt.Stringer {
	def, ok := definitions.Get(uiName)
	if !ok {
		panic(fmt.Sprintf("No definition found for %s", uiName))
	}

	return def
}
