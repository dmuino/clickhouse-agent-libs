package chagent

import (
	"os"
	"text/template"
)

func GetTemplate(inputName string) *template.Template {
	tpl, err := template.ParseFiles(inputName)
	logger.CheckErr(err)
	return tpl
}

func GenerateConfig(tpl *template.Template, outputName string, config interface{}) {
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