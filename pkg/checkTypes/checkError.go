/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checkTypes

type CheckError struct {
	Path     string
	Message  string
	Location int
}

func NewCheckError(path string, message string, location int) *CheckError {
	checkError := &CheckError{
		Path:     path,
		Message:  message,
		Location: location,
	}
	return checkError
}
