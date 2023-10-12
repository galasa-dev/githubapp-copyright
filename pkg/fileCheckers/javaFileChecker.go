/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package fileCheckers

import (
	"regexp"

	"github.com/galasa-dev/githubapp-copyright/pkg/checkTypes"
)

type JavaFileChecker struct {
	javaCommentBlockPattern      *regexp.Regexp
	javaCopyrightPattern         *regexp.Regexp
	javaExpectedCopyrightMessage string
}

func NewJavaFileChecker() FileChecker {
	this := new(JavaFileChecker)

	// \s means any whitespace character (including \n new lines)
	// [*] means a splat/star/asterisk character.
	// We are trying to all this:
	// A copyright message "Copyright contributors to the Galasa project" followed by
	// any number of lines with leading and trailing whitespace around an asterisk, followed by
	// a line containing <optional-whitespace>SPDX-License-Identifier:<optional-whitespace>EPL-2.0
	this.javaCopyrightPattern = regexp.MustCompile(`Copyright contributors to the Galasa project(\s*[*]\s*)*\s*[*]\s*SPDX-License-Identifier:\s*EPL-2[.]0`)

	this.javaCommentBlockPattern = regexp.MustCompile(`\s*\/[*]((.|\s)*)[*]\/`)

	this.javaExpectedCopyrightMessage = "\nExpected to see:\n/*\n * Copyright contributors to the Galasa project\n *\n * SPDX-License-Identifier: EPL-2.0\n */"

	return this
}

func (this *JavaFileChecker) CheckFileContent(content string, fileName string) *checkTypes.CheckError {

	var checkError *checkTypes.CheckError = nil

	commentBlockLocation := this.javaCommentBlockPattern.FindStringIndex(content)

	if commentBlockLocation == nil {
		checkError = &checkTypes.CheckError{
			Path:     fileName,
			Message:  "Did not find comment block." + this.javaExpectedCopyrightMessage,
			Location: 0,
		}
	} else {
		commentBlock := content[commentBlockLocation[0]:commentBlockLocation[1]]

		checkError = checkCommentBlock(&commentBlock, fileName, this.javaCopyrightPattern, this.javaExpectedCopyrightMessage)

		if checkError == nil {
			// last check,  the first comment block should be at the top
			if commentBlockLocation[0] != 0 {
				checkError = &checkTypes.CheckError{
					Path:     fileName,
					Message:  "Comment block containing copyright should be at the top of the file." + this.javaExpectedCopyrightMessage,
					Location: commentBlockLocation[0],
				}
			}
		}
	}

	return checkError
}
