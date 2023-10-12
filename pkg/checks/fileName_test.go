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

func TestNoExtensionReturnsBlank(t *testing.T) {
	fileExtension := extractFileExtension("nofileextension")
	assert.Equal(t, fileExtension, "")
}

func TestCanExtractJavaExtensionOk(t *testing.T) {
	fileExtension := extractFileExtension("myClassFile.java")
	assert.Equal(t, fileExtension, ".java")
}

func TestIgnoresMultipleDotsAndGetsRealFileExtension(t *testing.T) {
	fileExtension := extractFileExtension("myPackage.myClassFile.java")
	assert.Equal(t, fileExtension, ".java")
}
