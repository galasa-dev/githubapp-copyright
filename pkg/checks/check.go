/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type CheckError struct {
	Path     string
	Message  string
	Location int
}

type Checker interface {
	CheckPullRequest(webhook *Webhook, checkId int, pullRequestUrl string) (*[]CheckError, error)

	CheckFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) (*CheckError, error)

	UpdateCheckRun(webhook *Webhook, checkRunURL *string, errors *[]CheckError, fatalError *string) error
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
}

func NewChecker(tokenSupplier TokenSupplier) (*CheckerImpl, error) {

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

	return checker, err
}

func (checker *CheckerImpl) CheckPullRequest(webhook *Webhook, checkId int, pullRequestUrl string) (*[]CheckError, error) {
	log.Printf("(%v) Checking pullrequest '%v'", checkId, pullRequestUrl)

	var err error = nil
	installationId := webhook.Installation.Id

	checkErrors := make([]CheckError, 0)

	var token string
	token, err = checker.tokenSupplier.GetToken(installationId)

	if err != nil {

		client := &http.Client{}

		// Retrieve list of files
		for page := 1; ; page++ {
			filesUrl := fmt.Sprintf("%v/files?page=%v", pullRequestUrl, page)

			var req *http.Request
			req, err = http.NewRequest("GET", filesUrl, nil)
			if err != nil {
				return nil, err
			}

			req.Header.Add("Authorization", "Bearer "+token)
			req.Header.Add("Accept", "application/vnd.github.v3+json")

			var resp *http.Response
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}

			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				break
			}

			var bodyBytes []byte
			bodyBytes, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var files []File

			err = json.Unmarshal(bodyBytes, &files)
			if err != nil {
				return nil, err
			}

			if files == nil || len(files) < 1 {
				break
			}

			for _, file := range files {
				var newError *CheckError
				newError, err = checker.CheckFile(webhook, checkId, &token, client, &file)
				if err == nil {
					if newError != nil {
						log.Printf("(%v) Found problem with file %v - %v", checkId, file.Filename, newError.Message)
						checkErrors = append(checkErrors, *newError)
					}
				} else {
					break
				}
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

func (checker *CheckerImpl) CheckFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) (*CheckError, error) {

	var err error = nil

	// we dont care about deleted files
	if file.Status == "removed" {
		return nil, err
	}

	// Check for Java files
	if strings.HasSuffix(file.Filename, ".java") {
		return checker.checkJavaFile(webhook, checkId, token, client, file)
	}

	// Check for Go files, same as java
	if strings.HasSuffix(file.Filename, ".go") {
		return checker.checkJavaFile(webhook, checkId, token, client, file)
	}

	// Check for Typescript files, same as Java
	if strings.HasSuffix(file.Filename, ".ts") || strings.HasSuffix(file.Filename, ".tsx") {
		return checker.checkJavaFile(webhook, checkId, token, client, file)
	}

	// Check for JavaScript files, same as Java
	if strings.HasSuffix(file.Filename, ".js") {
		return checker.checkJavaFile(webhook, checkId, token, client, file)
	}

	// Check for Yaml files
	if strings.HasSuffix(file.Filename, ".yaml") {
		return checker.checkYamlFile(webhook, checkId, token, client, file)
	}

	// Check for Bash Script files, same as Yaml
	if strings.HasSuffix(file.Filename, ".sh") {
		return checker.checkYamlFile(webhook, checkId, token, client, file)
	}

	// Not a file we are concerned about
	return nil, err
}

func (checker *CheckerImpl) checkJavaFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) (*CheckError, error) {
	log.Printf("(%v) Checking file %v - %v\n", checkId, file.Filename, file.Sha)
	var checkError *CheckError = nil
	var content string
	var err error = nil
	content, err = checker.getFileContent(token, client, &file.ContentsURL)
	if err != nil {
		fatalMessage := fmt.Sprintf("Failed to access the content of the file for checking - %v", err)
		checkError = &CheckError{
			Path:     file.Filename,
			Message:  fatalMessage,
			Location: 0,
		}
	} else {
		checkError = checker.checkJavaFileContent(content, file.Filename)
	}

	return checkError, err
}

func (checker *CheckerImpl) checkJavaFileContent(content string, fileName string) *CheckError {

	var checkError *CheckError = nil
	var fileType = "java"

	commentBlockLocation := checker.javaCommentBlockPattern.FindStringIndex(content)

	if commentBlockLocation == nil {
		checkError = &CheckError{
			Path:     fileName,
			Message:  "Did not find comment block." + checker.javaExpectedCopyrightMessage,
			Location: 0,
		}
	} else {
		commentBlock := content[commentBlockLocation[0]:commentBlockLocation[1]]

		checkError = checker.checkCommentBlock(&commentBlock, fileName, fileType)

		if checkError == nil {
			// last check,  the first comment block should be at the top
			if commentBlockLocation[0] != 0 {
				checkError = &CheckError{
					Path:     fileName,
					Message:  "Comment block containing copyright should be at the top of the file." + checker.javaExpectedCopyrightMessage,
					Location: commentBlockLocation[0],
				}
			}
		}
	}

	return checkError
}

func (checker *CheckerImpl) checkYamlFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) (*CheckError, error) {
	log.Printf("(%v) Checking file %v - %v\n", checkId, file.Filename, file.Sha)

	var checkError *CheckError = nil
	var err error = nil
	var content string
	content, err = checker.getFileContent(token, client, &file.ContentsURL)
	if err != nil {
		fatalMessage := fmt.Sprintf("Failed to access the content of the file for checking - %v", err)
		checkError = &CheckError{
			Path:     file.Filename,
			Message:  fatalMessage,
			Location: 0,
		}
	} else {
		checkError = checker.checkYamlFileContent(content, file.Filename)
	}

	return checkError, err
}

func (checker *CheckerImpl) checkYamlFileContent(content string, fileName string) *CheckError {
	var checkError *CheckError = nil
	var fileType = "yaml"

	commentBlock := ""

	//if it is a bash script (.sh)
	//ignore the first line that starts with !#
	//and any subsequent whitespace
	if strings.HasSuffix(fileName, ".sh") {
		nextLine := strings.Index(content, "\n")
		content = strings.TrimSpace(content[nextLine:])
	}

	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			break
		}
		commentBlock = commentBlock + line + "\n"
	}

	// check we have a comment block at the begining of the file
	if commentBlock == "" {
		checkError = &CheckError{
			Path:     fileName,
			Message:  "A comment block is missing at the start of the file." + checker.hashExpectedCopyrightMessage,
			Location: 0,
		}
	} else {
		checkError = checker.checkCommentBlock(&commentBlock, fileName, fileType)
	}

	return checkError
}

func (checker *CheckerImpl) checkCommentBlock(commentBlock *string, fileName string, fileType string) *CheckError {
	var checkError *CheckError = nil
	var copyrights [][]int
	var expectedCopyrightMessage string

	// Check to see if it has the copyright text
	if fileType == "java" {
		copyrights = checker.javaCopyrightPattern.FindAllStringSubmatchIndex(*commentBlock, -1)
		expectedCopyrightMessage = checker.javaExpectedCopyrightMessage
	} else if fileType == "yaml" {
		copyrights = checker.hashCopyrightPattern.FindAllStringSubmatchIndex(*commentBlock, -1)
		expectedCopyrightMessage = checker.hashExpectedCopyrightMessage
	}

	if len(copyrights) <= 0 {
		checkError = &CheckError{
			Path:     fileName,
			Message:  "Did not find copyright text in first comment block." + expectedCopyrightMessage,
			Location: 0,
		}
	}

	if len(copyrights) > 1 {
		checkError = &CheckError{
			Path:     fileName,
			Message:  "Found too many copyright texts in first comment block" + expectedCopyrightMessage,
			Location: 0,
		}
	}

	return checkError
}

func (checker *CheckerImpl) getFileContent(token *string, client *http.Client, contentURL *string) (string, error) {
	contents := ""

	var err error = nil
	var req *http.Request
	req, err = http.NewRequest("GET", *contentURL, nil)
	if err == nil {

		req.Header.Add("Authorization", "Bearer "+*token)
		req.Header.Add("Accept", "application/vnd.github.v3.raw")
		var resp *http.Response
		resp, err = client.Do(req)
		if err == nil {

			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				err = errors.New("invalid response from content fetch " + resp.Status)
			} else {

				var bodyBytes []byte
				bodyBytes, err = io.ReadAll(resp.Body)
				if err == nil {

					contents = string(bodyBytes)
				}
			}
		}
	}

	return contents, err
}

// Create a 'check run' on github.
func (checker *CheckerImpl) CreateCheckRun(webhook *Webhook, headSha *string) (*string, error) {

	var url *string = nil

	installationId := webhook.Installation.Id

	var err error = nil
	var token string
	token, err = checker.tokenSupplier.GetToken(installationId)
	if err == nil {

		client := &http.Client{}

		checkRun := CheckRun{
			Name:    "copyright",
			HeadSha: headSha,
			Status:  "in_progress",
			Output: CheckRunOutput{
				Title:   "Galasa copyright check",
				Summary: "Checks for updated copyright years and licence text",
			},
		}

		var checkRunBytes []byte
		checkRunBytes, err = json.Marshal(&checkRun)
		if err == nil {

			// Post a status back to github.
			var req *http.Request
			req, err = http.NewRequest("POST", webhook.Repository.RepositoryURL+"/check-runs", bytes.NewReader(checkRunBytes))
			if err == nil {

				req.Header.Add("Authorization", "Bearer "+token)
				req.Header.Add("Accept", "application/vnd.github.v3+json")
				req.Header.Add("Content-Type", "application/vnd.github.v3+json")

				var resp *http.Response
				resp, err = client.Do(req)
				if err == nil {

					defer resp.Body.Close()

					if resp.StatusCode != 201 {
						err = errors.New(fmt.Sprintf("Got a non-201 status code from a POST github. status code=%d", resp.StatusCode))
					} else {
						var bodyBytes []byte
						bodyBytes, err = io.ReadAll(resp.Body)
						if err == nil {

							var response CheckRun

							err = json.Unmarshal(bodyBytes, &response)
							if err == nil {
								url = response.Url
							}
						}
					}
				}
			}
		}
	}

	return url, err

}

// Update the status of a previously-created 'check run' which exists at the end of a URL in github.
func (checker *CheckerImpl) UpdateCheckRun(webhook *Webhook, checkRunURL *string, checkErrors *[]CheckError, fatalError *string) error {

	var err error = nil
	var token string

	token, err = checker.tokenSupplier.GetToken(webhook.Installation.Id)
	if err == nil {

		client := &http.Client{}

		checkRun := CheckRun{
			Name:   "copyright",
			Status: "completed",
			Output: CheckRunOutput{
				Title:   "Galasa copyright check",
				Summary: "Checks for updated copyright years and licence text",
			},
		}

		conclusion := "success"

		if fatalError != nil {
			conclusion = "failure"
			checkRun.Output.Summary = *fatalError
		} else if len(*checkErrors) > 0 {
			conclusion = "failure"
			annotations := make([]CheckRunAnnotation, 0)

			for _, error := range *checkErrors {
				annotation := CheckRunAnnotation{
					Path:      error.Path,
					Message:   error.Message,
					Level:     "failure",
					StartLine: 1,
					EndLine:   1,
				}
				annotations = append(annotations, annotation)
			}

			checkRun.Output.Annotations = &annotations
		}
		checkRun.Conclusion = &conclusion

		var checkRunBytes []byte
		checkRunBytes, err = json.Marshal(&checkRun)
		if err == nil {

			var req *http.Request
			req, err = http.NewRequest("PATCH", *checkRunURL, bytes.NewReader(checkRunBytes))
			if err == nil {

				req.Header.Add("Authorization", "Bearer "+token)
				req.Header.Add("Accept", "application/vnd.github.v3+json")
				req.Header.Add("Content-Type", "application/vnd.github.v3+json")

				var resp *http.Response
				resp, err = client.Do(req)
				if err == nil {

					defer resp.Body.Close()

					var bodyBytes []byte
					bodyBytes, err = io.ReadAll(resp.Body)
					if err == nil {

						data := string(bodyBytes)

						if resp.StatusCode != 200 {
							log.Fatalf("Fatal error - %v", data)
							err = errors.New(fmt.Sprintf("Non-200 status returned from github. %d", resp.StatusCode))
						}
					}
				}
			}

		}
	}

	if err != nil {
		log.Fatalf("Fatal error - %v", err)
	}

	return err
}
