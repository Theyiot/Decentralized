package util

import (
	"os"
)

func FailOnError(err error) {
	if err != nil {
		println("Error: ", err.Error())
		os.Exit(0)
	}
}

func CheckAndPrintError(err error) bool {
	if err != nil {
		println("ERROR : ", err.Error())
	}
	return err != nil
}
