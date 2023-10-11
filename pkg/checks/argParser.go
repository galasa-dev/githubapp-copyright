/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"errors"
	"fmt"
)

type FieldValuesParsed struct {
	GithubAuthKeyFilePath string
	IsDebugEnabled        bool
}

type CommandLineArgParser interface {
	Parse() (*FieldValuesParsed, error)
}

type CommandLineArgParserImpl struct {
	argSequence StringIterator
	console     Console
}

const (
	COMMAND_FLAG_GITHUB_AUTH_KEY_FILE = "--githubAuthKeyFile"
	COMMAND_FLAG_DEBUG                = "--debug"
)

func NewCommandLineArgParserImpl(args []string, console Console) (CommandLineArgParser, error) {
	var err error = nil
	parser := new(CommandLineArgParserImpl)

	parser.argSequence, err = NewStringIterator(args)
	parser.console = console

	return parser, err
}

func (this *CommandLineArgParserImpl) Parse() (*FieldValuesParsed, error) {
	var err error = nil
	results := new(FieldValuesParsed)

	// Skip over the first arg, it's the command which called this program.
	arg, isDone := this.argSequence.Next()

	for {

		arg, isDone = this.argSequence.Next()
		if isDone {
			// Finished parsing parameters. None left to look at.
			break
		}

		switch arg {
		case COMMAND_FLAG_GITHUB_AUTH_KEY_FILE:
			{
				arg, isDone := this.argSequence.Next()
				if isDone {
					// Ran out of args, expected a value.
					msg := fmt.Sprintf("Error: Flag %s requires a value.\n", COMMAND_FLAG_GITHUB_AUTH_KEY_FILE)
					err = errors.New(msg)
					this.console.Write(msg)
				} else {
					results.GithubAuthKeyFilePath = arg
				}
			}

		case COMMAND_FLAG_DEBUG:
			{
				results.IsDebugEnabled = true
			}

		default:
			msg := fmt.Sprintf("Error: Unrecognised parameter '%s'\n", arg)
			err = errors.New(msg)
			this.console.Write(msg)
		}

		if err != nil {
			break
		}
	}

	if results.GithubAuthKeyFilePath == "" {
		results.GithubAuthKeyFilePath = "key.pem"
	}

	return results, err
}
