/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"strings"
)

func extractFileExtension(fileName string) string {
	fileExtension := ""

	indexOfLastDot := strings.LastIndex(fileName, ".")
	if indexOfLastDot >= 0 {
		fileExtension = fileName[indexOfLastDot:]
	}

	return fileExtension
}
