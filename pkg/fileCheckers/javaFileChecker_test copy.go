/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package fileCheckers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckJavaContentFindsCopyrightOk(t *testing.T) {
	// Given

	checker := NewJavaFileChecker()
	var content = `/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
`
	var fileName = "test.java"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckJavaContentFindsCopyrightMissing(t *testing.T) {
	// Given
	checker := NewJavaFileChecker()
	var content = `/*
 *
 *
 */
`
	var fileName = "test.java"
	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find copyright text in first comment block")
}

func TestCheckJavaContentFindsNoComment(t *testing.T) {
	// Given
	checker := NewJavaFileChecker()
	var content = `Hello, world!
`
	var fileName = "test.java"
	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find comment block")
}

func TestCheckJavaContentFindsLicenseMissing(t *testing.T) {
	// Given
	var content = `/*
 * Copyright contributors to the Galasa project
 */
`
	var fileName = "test.java"
	// When..
	checker := NewJavaFileChecker()
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find copyright text in first comment block")
}

// should copyright comments only be at the top?
func TestCheckJavaContentFindsCopyrightOkAndHasLeadingText(t *testing.T) {
	// Given
	checker := NewJavaFileChecker()
	var content = `leading text here
	and more leading text
/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
`
	var fileName = "test.java"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Comment block containing copyright should be at the top of the file")
}

func TestCheckJavaContentFindsLicenseMissingAndHasLeadingText(t *testing.T) {
	// Given
	checker := NewJavaFileChecker()
	var content = `leading text here
and more leading text
/*
 * Copyright contributors to the Galasa project
 */
`
	var fileName = "test.java"
	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find copyright text in first comment block")
}

func TestCheckJavaContentFindsTooManyCopyright(t *testing.T) {
	// Given
	checker := NewJavaFileChecker()
	var content = `/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 *
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
`
	var fileName = "test.java"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Found too many copyright texts in first comment block")
}

func TestCheckJavaContentFindsCopyrightCommentJoinedWithAnotherComment(t *testing.T) {
	// Given
	checker := NewJavaFileChecker()
	var content = `/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 *
 * Another comment here
 */
`
	var fileName = "test.java"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckJavaContentFindsCopyrightOkWhenOtherCommentsArePresent(t *testing.T) {
	// Given
	checker := NewJavaFileChecker()
	var content = `/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */

 /*
 * don't detect me!
 */
`
	var fileName = "test.java"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}
