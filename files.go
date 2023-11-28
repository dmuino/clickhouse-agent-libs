package chagent

import "os"

func CopyFile(inputName string, outputName string) {
	logger := GetLogger("CopyFile")
	input, err := os.ReadFile(inputName)
	logger.CheckErr(err)
	err = os.WriteFile(outputName, input, 0644)
	logger.CheckErr(err)
}
