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

func TestCanCreateConsole(t *testing.T) {
	console, err := NewConsole()
	assert.Nil(t, err)
	assert.NotNil(t, console)
}
