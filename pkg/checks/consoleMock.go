/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import "strings"

type ConsoleMock struct {
	messages []string
}

func NewConsoleMock() *ConsoleMock {
	console := new(ConsoleMock)
	return console
}

func (this *ConsoleMock) Write(message string) {
	this.messages = append(this.messages, message)
}

func (this *ConsoleMock) getMessages() []string {
	return this.messages
}

func (this *ConsoleMock) contains(subString string) bool {
	isFound := false
	for _, msg := range this.messages {
		if strings.Contains(msg, subString) {
			isFound = true
			break
		}
	}
	return isFound
}

func (this *ConsoleMock) getAllAsString() string {
	result := ""
	for _, msg := range this.messages {
		result = result + msg
	}
	return result
}
