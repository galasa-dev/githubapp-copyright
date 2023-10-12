/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import _ "embed"

//go:embed resources/version.txt
var versionText string

func GetVersion() string {
	return versionText
}
