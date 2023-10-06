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

func TestCanCreateAStringIterator(t *testing.T) {
	args := []string{"copyright"}
	iter, err := NewStringIterator(args)
	assert.Nil(t, err)
	assert.NotNil(t, iter)
}

func TestCanWalkOneStringInList(t *testing.T) {
	args := []string{"copyright"}
	iter, _ := NewStringIterator(args)
	item, isDone := iter.Next()
	assert.Equal(t, "copyright", item)
	assert.False(t, isDone)
}

func TestGetIsDoneTrueIfWalkBeyondListEnd(t *testing.T) {
	args := []string{"copyright"}
	iter, _ := NewStringIterator(args)
	item, isDone := iter.Next()

	item, isDone = iter.Next()
	assert.Equal(t, "", item)
	assert.True(t, isDone)
}
