/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

type StringIterator interface {
	// Gets the next string from the input slice. If we run out of input, we return ""
	Next() (string, bool)
}

type StringIteratorImpl struct {
	args                  []string
	nextArgToProcessIndex int
}

func NewStringIterator(args []string) (StringIterator, error) {
	var err error = nil
	this := new(StringIteratorImpl)

	this.args = args

	this.nextArgToProcessIndex = 0

	return this, err
}

func (this *StringIteratorImpl) Next() (string, bool) {
	var next string
	isDone := true
	if this.args != nil {
		if this.nextArgToProcessIndex >= len(this.args) {
			next = ""
		} else {
			next = this.args[this.nextArgToProcessIndex]
			this.nextArgToProcessIndex += 1
			isDone = false
		}
	}
	return next, isDone
}
