/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"log"
	"regexp"

	"github.com/galasa-dev/githubapp-copyright/pkg/checkTypes"
	"github.com/galasa-dev/githubapp-copyright/pkg/fileCheckers"
)

type Checker interface {
	CheckFilesChanged(token string, url string) ([]checkTypes.CheckError, error)

	CheckFile(token string, file *File) *checkTypes.CheckError
}

type CheckerImpl struct {
	javaCommentBlockPattern *regexp.Regexp

	//licencePattern   *regexp.Regexp
	javaCopyrightPattern         *regexp.Regexp
	hashCopyrightPattern         *regexp.Regexp
	javaExpectedCopyrightMessage string
	hashExpectedCopyrightMessage string

	// The index is the file extension (including the dot) eg: ".java"
	// The value is the file checker which will be used.
	checkersByExtension map[string]fileCheckers.FileChecker

	gitHubClient GitHubClient
}

func NewChecker(client GitHubClient) (Checker, error) {

	var err error = nil

	checker := new(CheckerImpl)

	checker.gitHubClient = client
	checker.javaCommentBlockPattern = regexp.MustCompile(`\s*\/[*]((.|\s)*)[*]\/`)

	// \s means any whitespace character (including \n new lines)
	// [*] means a splat/star/asterisk character.
	// We are trying to all this:
	// A copyright message "Copyright contributors to the Galasa project" followed by
	// any number of lines with leading and trailing whitespace around an asterisk, followed by
	// a line containing <optional-whitespace>SPDX-License-Identifier:<optional-whitespace>EPL-2.0
	checker.javaCopyrightPattern = regexp.MustCompile(`Copyright contributors to the Galasa project(\s*[*]\s*)*\s*[*]\s*SPDX-License-Identifier:\s*EPL-2[.]0`)
	checker.hashCopyrightPattern = regexp.MustCompile(`Copyright contributors to the Galasa project(\s*[#]\s*)*\s*[#]\s*SPDX-License-Identifier:\s*EPL-2[.]0`)

	checker.javaExpectedCopyrightMessage = "\nExpected to see:\n/*\n * Copyright contributors to the Galasa project\n *\n * SPDX-License-Identifier: EPL-2.0\n */"
	checker.hashExpectedCopyrightMessage = "\nExpected to see:\n#\n# Copyright contributors to the Galasa project\n#\n# SPDX-License-Identifier: EPL-2.0\n#"

	var javaChecker fileCheckers.FileChecker
	javaChecker = fileCheckers.NewJavaFileChecker()

	var yamlChecker fileCheckers.FileChecker
	yamlChecker = fileCheckers.NewYamlFileChecker()

	checker.checkersByExtension = map[string]fileCheckers.FileChecker{
		".java": javaChecker,
		".go":   javaChecker,
		".ts":   javaChecker,
		".tsx":  javaChecker,
		".js":   javaChecker,
		".yaml": yamlChecker,
		".sh":   yamlChecker,
	}

	return checker, err
}

func (this *CheckerImpl) CheckFilesChanged(token string, url string) ([]checkTypes.CheckError, error) {
	var allFiles []File
	var err error = nil

	var checkErrors []checkTypes.CheckError = make([]checkTypes.CheckError, 0)

	allFiles, err = this.gitHubClient.GetFilesChanged(token, url)

	for _, file := range allFiles {
		var newCheckError *checkTypes.CheckError
		newCheckError = this.CheckFile(token, &file)

		if newCheckError != nil {
			log.Printf("Found problem with file %v - %v", file.Filename, newCheckError.Message)
			checkErrors = append(checkErrors, *newCheckError)
		}

		// Continue to check the next file also.
	}
	return checkErrors, err

}

func (this *CheckerImpl) CheckFile(token string, file *File) *checkTypes.CheckError {

	var err error = nil
	var checkError *checkTypes.CheckError

	// we dont care about deleted files
	if file.Status == "removed" {
		return nil
	}

	fileExtension := extractFileExtension(file.Filename)

	// Decide which file checker we want to use.
	fileChecker, isExtensionRecognised := this.checkersByExtension[fileExtension]

	if !isExtensionRecognised {
		// Don't bother getting the file if we don't know how to check it for copyright.
		log.Printf("File file %s is not checked because extension %s is not checked for copyright.\n", file.Filename, fileExtension)
	} else {

		var fileContent string
		fileContent, err = this.gitHubClient.GetFileContentFromGithub(token, file)
		if err == nil {

			checkError = fileChecker.CheckFileContent(fileContent, file.Filename)
		} else {
			// Turn the error into a checker error so it fails the check in github.
			log.Printf("Failed to check file %s. Reason: %s\n", file.Filename, err.Error())
			checkError = &checkTypes.CheckError{
				Path:     file.Filename,
				Message:  err.Error(),
				Location: 0,
			}
		}
	}

	return checkError
}
