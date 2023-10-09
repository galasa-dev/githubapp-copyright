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

func checkCommentBlock(commentBlock *string, fileName string, copyrightPattern *regexp.Regexp, expectedCopyrightMessage string) *checkTypes.CheckError {
	var checkError *checkTypes.CheckError = nil
	var copyrights [][]int

	// Check to see if it has the copyright text
	copyrights = copyrightPattern.FindAllStringSubmatchIndex(*commentBlock, -1)

	if len(copyrights) <= 0 {
		checkError = checkTypes.NewCheckError(
			fileName,
			"Did not find copyright text in first comment block."+expectedCopyrightMessage,
			0,
		)
	}

	if len(copyrights) > 1 {
		checkError = &checkTypes.CheckError{
			Path:     fileName,
			Message:  "Found too many copyright texts in first comment block" + expectedCopyrightMessage,
			Location: 0,
		}
	}

	return checkError
}
