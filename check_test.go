/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	//"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckJavaContentFindsCopyrightOk(t *testing.T) {
	// Given
	var content = `/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
`
	var fileName = "test.java"

	// When..
	checkError := checkJavaFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckJavaContentFindsCopyrightMissing(t *testing.T) {
	// Given
	var content = `/*
 * 
 *
 */
`
	var fileName = "test.java"

	// When..
	checkError := checkJavaFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}
