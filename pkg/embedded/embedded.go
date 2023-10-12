/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	_ "embed"
)

//go:embed resources/version.txt
var versionText string

//go:embed resources/build-date.txt
var buildDate string

//go:embed resources/git-commit-sha.txt
var gitCommitSha string

//go:embed resources/copyright.txt
var copyrightText string

func GetVersion() string {
	return versionText
}

func GetBuildDate() string {
	return buildDate
}

func GetLatestGitCommitSha() string {
	return gitCommitSha
}

func GetCopyright() string {
	return copyrightText
}
