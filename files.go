package chagent

import "os"

func CopyFile(inputName string, outputName string) {
	logger := GetLogger("CopyFile")
	input, err := os.ReadFile(inputName)
	logger.CheckErr(err)
	err = os.WriteFile(outputName, input, 0644)
	logger.CheckErr(err)
}

func FileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func TouchFile(name string) {
	if FileExists(name) {
		return
	}
	// ignore errors
	file, err := os.Create(name)
	if err != nil {
		_ = file.Close()
	}
}
