/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"os"
)

type Console interface {
	Write(string)
}

type ConsoleImpl struct {
}

func NewConsole() (Console, error) {
	var err error = nil
	console := new(ConsoleImpl)
	return console, err
}

func (*ConsoleImpl) Write(message string) {
	os.Stdout.WriteString(message)
}
