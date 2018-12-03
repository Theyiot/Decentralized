package util

import "os"

/*
	Checks whether there is an error and stop the program if there is one. This method should only be used
	for fatal errors, since it stops the execution of the program
 */
func FailOnError(err error) {
	if err != nil {
		println("Error: ", err.Error())
		os.Exit(0)
	}
}

/*
	Checks whether there is an error and printl it if there is one. This method also returns true if the is an
	error and false otherwise
 */
func CheckAndPrintError(err error) bool {
	if err != nil {
		println("ERROR : ", err.Error())
	}
	return err != nil
}
