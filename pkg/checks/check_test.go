/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
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
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find copyright text in first comment block")
}

func TestCheckJavaContentFindsLicenseMissing(t *testing.T) {
	// Given
	var content = `/*
 * Copyright contributors to the Galasa project
 */
`
	var fileName = "test.java"
	// When..
	checkError := checkJavaFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find copyright text in first comment block")
}
