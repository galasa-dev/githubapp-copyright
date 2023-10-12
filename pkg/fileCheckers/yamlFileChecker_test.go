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

func TestCheckYamlContentFindsCopyrightOk(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckYamlContentFindsNoComment(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()

	var content = `Hello, World!
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "A comment block is missing at the start of the file")
}

func TestCheckYamlContentFindsCopyrightMissing(t *testing.T) {

	// Given
	checker := NewYamlFileChecker()
	var content = `# Hello, world!
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find copyright text in first comment block")
}

func TestCheckYamlContentFindsLicenseMissing(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#
# Copyright contributors to the Galasa project
#
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find copyright text in first comment block")
}

// should copyright comments only be at the top?
func TestCheckYamlContentFindsCopyrightOkWithLeadingText(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `leading text
and more leading text

#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "A comment block is missing at the start of the file")
}

func TestCheckYamlContentFindsTooManyCopyright(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Found too many copyright texts in first comment block")
}

func TestCheckYamlContentFindsCopyrightCommentJoinedWithAnotherComment(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
# Another comment here
#
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckYamlContentFindsCopyrightOkWhenOtherCommentsArePresent(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#

#
# Don't detect me!
#
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckYamlContentFindsCopyrightWithNoLeadingOrEndingHashes(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0

#don't detect me!
`
	var fileName = "test.yaml"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckBashContentCopyrightOk(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#! bin/bash

#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
`
	var fileName = "test.sh"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckBashContentFindsCopyrightMissing(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#! bin
#
# SPDX-License-Identifier: EPL-2.0
#
`
	var fileName = "test.sh"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Did not find copyright text in first comment block")
}

// should copyright comments only be at the top?
func TestCheckBashContentWithLeadingCommentOrText(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#! bin/bin
leading text
and more leading text

#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
`
	var fileName = "test.sh"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "A comment block is missing at the start of the file")
}

func TestCheckBashContentFindsNoComment(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#! bin
Hello, World!
`
	var fileName = "test.sh"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "A comment block is missing at the start of the file")
}

func TestCheckBashContentFindsTooManyCopyright(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#! bin
#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
`
	var fileName = "test.sh"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.NotNil(t, checkError)
	assert.Contains(t, checkError.Message, "Found too many copyright texts in first comment block")
}

func TestCheckBashContentFindsCopyrightOkWhenOtherCommentsArePresent(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#! bin
#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#

#
# Don't detect me!
#
`
	var fileName = "test.sh"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestBashYamlContentFindsCopyrightWithNoLeadingOrEndingHashes(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#! bash
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0

#don't detect me!
`
	var fileName = "test.sh"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}

func TestCheckBashContentFindsCopyrightCommentJoinedWithAnotherComment(t *testing.T) {
	// Given
	checker := NewYamlFileChecker()
	var content = `#! bin
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
# Another comment here
#
`
	var fileName = "test.sh"

	// When..
	checkError := checker.CheckFileContent(content, fileName)

	// Then...
	assert.Nil(t, checkError)
}
