/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package fileCheckers

import "github.com/galasa-dev/githubapp-copyright/pkg/checkTypes"

type FileChecker interface {
	CheckFileContent(content string, fileName string) *checkTypes.CheckError
}
