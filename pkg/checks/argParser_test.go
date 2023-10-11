/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanCreateAParser(t *testing.T) {
	args := []string{"copyright"}
	var console Console
	console, _ = NewConsole()
	parser, err := NewCommandLineArgParserImpl(args, console)
	assert.Nil(t, err)
	assert.NotNil(t, parser)
}

func TestNoArgsReturnsValuesStructure(t *testing.T) {
	args := []string{"copyright"}
	var console Console
	console, _ = NewConsole()
	parser, err := NewCommandLineArgParserImpl(args, console)
	values, err := parser.Parse()
	assert.Nil(t, err)
	assert.NotNil(t, values)
}

func TestNoArgsReturnsValuesDefaultingGithubAuthKeyFilePath(t *testing.T) {
	args := []string{"copyright"}
	var console Console
	console, _ = NewConsole()
	parser, err := NewCommandLineArgParserImpl(args, console)
	values, err := parser.Parse()
	assert.Nil(t, err)
	assert.NotNil(t, values)

	assert.Equal(t, "key.pem", values.GithubAuthKeyFilePath)
}

func TestCanSpecifyGithubAuthKeyFilePath(t *testing.T) {
	args := []string{"copyright", "--githubAuthKeyFile", "myFilePath"}

	console := NewConsoleMock()
	parser, err := NewCommandLineArgParserImpl(args, console)
	values, err := parser.Parse()
	assert.Nil(t, err)
	assert.NotNil(t, values)
	assert.Equal(t, "myFilePath", values.GithubAuthKeyFilePath)
}

func TestGithubAuthKeyFilePathFlagWithNoValueGivesError(t *testing.T) {
	args := []string{"copyright", "--githubAuthKeyFile"}

	console := NewConsoleMock()
	parser, _ := NewCommandLineArgParserImpl(args, console)
	_, err := parser.Parse()
	assert.NotNil(t, err)
	assert.Contains(t, console.getAllAsString(), fmt.Sprintf("Error: Flag %s requires a value.\n", COMMAND_FLAG_GITHUB_AUTH_KEY_FILE))
}

func TestUnknownParamGivesError(t *testing.T) {
	args := []string{"copyright", "--githubAuthKeyFile", "myFilePath", "garbage"}

	console := NewConsoleMock()
	parser, err := NewCommandLineArgParserImpl(args, console)
	values, err := parser.Parse()
	assert.NotNil(t, err)
	assert.True(t, console.contains("Error: Unrecognised parameter 'garbage'"))
	assert.Equal(t, "myFilePath", values.GithubAuthKeyFilePath)
}

func TestDebugFlagPresentParsesAsDebugEnabled(t *testing.T) {
	args := []string{"copyright", "--githubAuthKeyFile", "myFilePath", "--debug"}

	console := NewConsoleMock()
	parser, err := NewCommandLineArgParserImpl(args, console)
	values, err := parser.Parse()
	assert.Nil(t, err)
	assert.NotNil(t, values)
	assert.Equal(t, "myFilePath", values.GithubAuthKeyFilePath)
	assert.Equal(t, true, values.IsDebugEnabled)
}

func TestDebugFlagMissingParsesAsDebugFalse(t *testing.T) {
	args := []string{"copyright", "--githubAuthKeyFile", "myFilePath"}

	console := NewConsoleMock()
	parser, err := NewCommandLineArgParserImpl(args, console)
	values, err := parser.Parse()
	assert.Nil(t, err)
	assert.NotNil(t, values)
	assert.Equal(t, false, values.IsDebugEnabled)
}
