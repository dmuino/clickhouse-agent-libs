package chagent

import (
	"bytes"
	"os"
	"text/template"
)

func GetTemplate(inputName string) *template.Template {
	tpl, err := template.ParseFiles(inputName)
	GetLogger("GetTemplate").CheckErr(err)
	return tpl
}

func GenerateConfig(tpl *template.Template, outputName string, config interface{}) {
	logger := GetLogger("GenerateConfig")
	tmpName := outputName + ".tmp"
	f, err := os.Create(tmpName)
	logger.CheckErr(err)
	err = tpl.Execute(f, config)
	logger.CheckErr(err)
	err = f.Close()
	logger.CheckErr(err)
	err = os.Rename(tmpName, outputName)
	logger.CheckErr(err)
}

func ExpandTemplate(tpl *template.Template, data interface{}) string {
	var output []byte
	writer := bytes.NewBuffer(output)
	err := tpl.Execute(writer, data)
	GetLogger("ExpandTemplate").CheckErr(err)
	return writer.String()
}
