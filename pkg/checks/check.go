/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"log"
	"net/http"
	"regexp"

	"github.com/galasa-dev/githubapp-copyright/pkg/checkTypes"
	"github.com/galasa-dev/githubapp-copyright/pkg/fileCheckers"
)

type Checker interface {
	CheckPullRequest(webhook *Webhook, checkId int, pullRequestUrl string) (*[]checkTypes.CheckError, error)

	CheckFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) *checkTypes.CheckError

	UpdateCheckRun(webhook *Webhook, checkRunURL *string, errors *[]checkTypes.CheckError, fatalError *string) error
	CreateCheckRun(webhook *Webhook, headSha *string) (*string, error)
}

type CheckerImpl struct {
	tokenSupplier TokenSupplier

	javaCommentBlockPattern *regexp.Regexp

	//licencePattern   *regexp.Regexp
	javaCopyrightPattern         *regexp.Regexp
	hashCopyrightPattern         *regexp.Regexp
	javaExpectedCopyrightMessage string
	hashExpectedCopyrightMessage string

	// The index is the file extension (including the dot) eg: ".java"
	// The value is the file checker which will be used.
	checkersByExtension map[string]fileCheckers.FileChecker
}

func NewChecker(tokenSupplier TokenSupplier) (Checker, error) {

	var err error = nil

	checker := new(CheckerImpl)

	checker.tokenSupplier = tokenSupplier

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

func (checker *CheckerImpl) CheckPullRequest(webhook *Webhook, checkId int, pullRequestUrl string) (*[]checkTypes.CheckError, error) {
	log.Printf("(%v) Checking pullrequest '%v'", checkId, pullRequestUrl)

	var err error = nil
	installationId := webhook.Installation.Id

	var checkErrors []checkTypes.CheckError = nil

	var token string
	token, err = checker.tokenSupplier.GetToken(installationId)

	if err != nil {

		var allFiles []File
		client := &http.Client{}

		allFiles, err = getFilesChangedByPullRequest(client, token, pullRequestUrl)

		for _, file := range allFiles {
			var newError *checkTypes.CheckError
			newError = checker.CheckFile(webhook, checkId, &token, client, &file)
			if newError != nil {
				log.Printf("(%v) Found problem with file %v - %v", checkId, file.Filename, newError.Message)
				checkErrors = append(checkErrors, *newError)
			}
		}

		if err == nil {
			if len(checkErrors) < 1 {
				return nil, nil
			}
		}
	}

	return &checkErrors, err
}

func (checker *CheckerImpl) CheckFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) *checkTypes.CheckError {

	var err error = nil
	var checkError *checkTypes.CheckError

	// we dont care about deleted files
	if file.Status == "removed" {
		return nil
	}

	fileExtension := extractFileExtension(file.Filename)

	// Decide which file checker we want to use.
	fileChecker, isExtensionRecognised := checker.checkersByExtension[fileExtension]

	if !isExtensionRecognised {
		// Don't bother getting the file if we don't know how to check it for copyright.
		log.Printf("File file %s is not checked because extension %s is not checked for copyright.\n", file.Filename, fileExtension)
	} else {

		var fileContent string
		fileContent, err = getFileContentFromGithub(token, client, file)
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
