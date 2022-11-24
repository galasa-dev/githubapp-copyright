/*
 * Copyright contributors to the Galasa project
 */
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	javaCommentBlockPattern *regexp.Regexp

	licencePattern   *regexp.Regexp
	copyrightPattern *regexp.Regexp
)

type CheckError struct {
	Path     string
	Message  string
	Location int
}

func init() {
	javaCommentBlockPattern = regexp.MustCompile("(?s)/\\*(.*?)\\*/")

	copyrightPattern = regexp.MustCompile("\\QCopyright contributors to the Galasa project\\E")
}

func checkPullRequest(webhook *Webhook, checkId int, pullRequestUrl string) (*[]CheckError, error) {
	log.Printf("(%v) Checking pullrequest '%v'", checkId, pullRequestUrl)

	installationId := webhook.Installation.Id

	token := getToken(installationId)

	client := &http.Client{}

	errors := make([]CheckError, 0)
	// Retrieve list of files
	for page := 1; ; page++ {
		filesUrl := fmt.Sprintf("%v/files?page=%v", pullRequestUrl, page)

		req, err := http.NewRequest("GET", filesUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Accept", "application/vnd.github.v3+json")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			break
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
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
			newError := checkFile(webhook, checkId, &token, client, &file)

			if newError != nil {
				log.Printf("(%v) Found problem with file %v - %v", checkId, file.Filename, newError.Message)
				errors = append(errors, *newError)
			}
		}
	}

	if len(errors) < 1 {
		return nil, nil
	}

	return &errors, nil
}

func checkFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) *CheckError {

	// we dont care about deleted files
	if file.Status == "removed" {
		return nil
	}

	// Check for Java files
	if strings.HasSuffix(file.Filename, ".java") {
		return checkJavaFile(webhook, checkId, token, client, file)
	}

	// Check for Go files, same as java
	if strings.HasSuffix(file.Filename, ".go") {
		return checkJavaFile(webhook, checkId, token, client, file)
	}

	// Check for Typescript files, same as Java
	if strings.HasSuffix(file.Filename, ".ts") {
		return checkJavaFile(webhook, checkId, token, client, file)
	}

	// Check for Yaml files
	if strings.HasSuffix(file.Filename, ".yaml") {
		return checkYamlFile(webhook, checkId, token, client, file)
	}

	// Not a file we are concerned about
	return nil
}

func checkJavaFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) *CheckError {
	log.Printf("(%v) Checking file %v - %v\n", checkId, file.Filename, file.Sha)
	content, err := getFileContent(token, client, &file.ContentsURL)
	if err != nil {
		fatalMessage := fmt.Sprintf("Failed to access the content of the file for checking - %v", err)
		return &CheckError{
			Path:     file.Filename,
			Message:  fatalMessage,
			Location: 0,
		}
	}

	commentBlockLocation := javaCommentBlockPattern.FindStringIndex(content)
	if commentBlockLocation == nil {
		return &CheckError{
			Path:     file.Filename,
			Message:  "Did not find comment block",
			Location: 0,
		}
	}

	commentBlock := content[commentBlockLocation[0]:commentBlockLocation[1]]

	error := checkCommentBlock(&commentBlock, file)
	if error != nil {
		return error
	}

	// last check,  the first comment block should be at the top
	if commentBlockLocation[0] != 0 {
		return &CheckError{
			Path:     file.Filename,
			Message:  fmt.Sprintf("Comment block containing copyright should be at the top. '%s'", copyrightPattern),
			Location: commentBlockLocation[0],
		}
	}

	// All is ok
	return nil
}

func checkYamlFile(webhook *Webhook, checkId int, token *string, client *http.Client, file *File) *CheckError {
	log.Printf("(%v) Checking file %v - %v\n", checkId, file.Filename, file.Sha)
	content, err := getFileContent(token, client, &file.ContentsURL)
	if err != nil {
		fatalMessage := fmt.Sprintf("Failed to access the content of the file for checking - %v", err)
		return &CheckError{
			Path:     file.Filename,
			Message:  fatalMessage,
			Location: 0,
		}
	}
	commentBlock := ""
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
		return &CheckError{
			Path:     file.Filename,
			Message:  "A comment block is missing at the start of the file",
			Location: 0,
		}
	}

	return checkCommentBlock(&commentBlock, file)
}

func checkCommentBlock(commentBlock *string, file *File) *CheckError {

	// Check to see if it has the copyright text

	copyrights := copyrightPattern.FindAllStringSubmatchIndex(*commentBlock, -1)

	if len(copyrights) <= 0 {
		return &CheckError{
			Path:     file.Filename,
			Message:  "Did not find copyright text in first comment block",
			Location: 0,
		}
	}

	if len(copyrights) > 1 {
		return &CheckError{
			Path:     file.Filename,
			Message:  "Found too many copyright texts in first comment block",
			Location: 0,
		}
	}

	// All is ok
	return nil
}

func getFileContent(token *string, client *http.Client, contentURL *string) (string, *error) {
	req, err := http.NewRequest("GET", *contentURL, nil)
	if err != nil {
		return "", &err
	}
	req.Header.Add("Authorization", "Bearer "+*token)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")
	resp, err := client.Do(req)
	if err != nil {
		return "", &err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		newError := errors.New("invalid response from content fetch " + resp.Status)
		return "", &newError
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", &err
	}

	return string(bodyBytes), nil
}

func createCheckRun(webhook *Webhook, headSha *string) *string {
	installationId := webhook.Installation.Id

	token := getToken(installationId)

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

	checkRunBytes, err := json.Marshal(&checkRun)
	if err != nil {
		panic(err) // TODO
	}

	req, err := http.NewRequest("POST", webhook.Repository.RepositoryURL+"/check-runs", bytes.NewReader(checkRunBytes))
	if err != nil {
		panic(err) // TODO
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Content-Type", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		panic(err) // TODO
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err) // TODO
	}

	if resp.StatusCode != 201 {
		panic(resp.StatusCode) // TODO
	}

	var response CheckRun

	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		panic(err) // TODO
	}

	return response.Url
}

func updateCheckRun(webhook *Webhook, checkRunURL *string, errors *[]CheckError, fatalError *string) {
	token := getToken(webhook.Installation.Id)

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
	} else if len(*errors) > 0 {
		conclusion = "failure"
		annotations := make([]CheckRunAnnotation, 0)

		for _, error := range *errors {
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

	checkRunBytes, err := json.Marshal(&checkRun)
	if err != nil {
		log.Fatalf("Fatal error - %v", err)
		return
	}

	req, err := http.NewRequest("PATCH", *checkRunURL, bytes.NewReader(checkRunBytes))
	if err != nil {
		log.Fatalf("Fatal error - %v", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Content-Type", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Fatal error - %v", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Fatal error - %v", err)
		return
	}

	data := string(bodyBytes)

	if resp.StatusCode != 200 {
		log.Fatalf("Fatal error - %v", data)
		return
	}
}
