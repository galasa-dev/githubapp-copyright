/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"log"
	"net/http"

	"github.com/galasa-dev/githubapp-copyright/pkg/checkTypes"
)

type GitHubClient interface {
	UpdateCheckRun(tokenSupplier TokenSupplier, webhook *Webhook, checkRunURL string, checkErrors []checkTypes.CheckError, fatalError string) error
	GetFilesChanged(token string, baseUrl string) ([]File, error)
	GetFileContentFromGithub(token string, file *File) (string, error)
	CreateCheckRun(tokenSupplier TokenSupplier, webhook *Webhook, headSha string) (string, error)
	GetNewToken(accessUrl string, githubAuthToken string) (tokenResponse InstallationToken, err error)
	LogHttpPayload(jsonBytes []byte)
}

type GitHubClientImpl struct {
	httpClient          *http.Client
	isHttpTrafficLogged bool
}

func NewGitHubClient(isHttpTrafficLogged bool) GitHubClient {
	this := new(GitHubClientImpl)
	this.httpClient = &http.Client{}
	this.isHttpTrafficLogged = isHttpTrafficLogged
	return this
}

func (this *GitHubClientImpl) LogHttpPayload(jsonBytes []byte) {
	if this.isHttpTrafficLogged {
		log.Printf("http payload:\n%s\n", string(jsonBytes))
	}
}

func (this *GitHubClientImpl) GetNewToken(accessUrl string, githubAuthToken string) (tokenResponse InstallationToken, err error) {

	var req *http.Request
	req, err = http.NewRequest("POST", accessUrl, nil)
	if err == nil {

		req.Header.Add("Authorization", "Bearer "+githubAuthToken)
		req.Header.Add("Accept", "application/vnd.github.v3+json")

		log.Printf("Sending HTTP POST to %s", accessUrl)

		var resp *http.Response
		resp, err = this.httpClient.Do(req)
		if err == nil {

			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				err = errors.New(fmt.Sprintf("Error. POST status code is not 201. Code: %v", resp.Status))
			} else {

				var buf bytes.Buffer
				_, err = io.Copy(&buf, resp.Body)
				if err == nil {

					var bodyBytes = buf.Bytes()

					err = json.Unmarshal(bodyBytes, &tokenResponse)
				}
			}
		}
	}
	return tokenResponse, err
}

func (this *GitHubClientImpl) GetFilesChanged(token string, baseUrl string) ([]File, error) {

	var err error = nil

	// Retrieve list of files
	var allFiles []File = make([]File, 0)

	// Keep asking for pages of results until we get an empty page or an error.
	for pageNumber := 1; ; pageNumber++ {

		var pageFiles []File
		pageFiles, err = this.getPageOfChangedFileNames(token, baseUrl, pageNumber)
		if err != nil {
			log.Printf("Page %d of file changes not obtained. %s", pageNumber, err.Error())
		} else {
			// Build the super-list of all files for this change
			allFiles = append(allFiles, pageFiles...)
		}

		if pageFiles == nil || len(pageFiles) < 1 {
			break
		}
	}
	return allFiles, err
}

func (this *GitHubClientImpl) getPageOfChangedFileNames(
	token string,
	baseUrl string,
	page int,
) ([]File, error) {

	var err error = nil
	var files []File
	filesUrl := fmt.Sprintf("%v/files?page=%v", baseUrl, page)

	var req *http.Request
	req, err = http.NewRequest("GET", filesUrl, nil)
	if err == nil {

		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Accept", "application/vnd.github.v3+json")

		log.Printf("Sending HTTP GET to %s", filesUrl)

		var resp *http.Response
		resp, err = this.httpClient.Do(req)
		if err == nil {

			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				err = errors.New(
					fmt.Sprintf(
						"Failed to get page %d of changed file names from %s. Return code was not OK. code=%v\n",
						page,
						baseUrl,
						resp.StatusCode,
					),
				)
			}

			var bodyBytes []byte
			bodyBytes, err = io.ReadAll(resp.Body)
			if err == nil {

				this.LogHttpPayload(bodyBytes)

				err = json.Unmarshal(bodyBytes, &files)
			}
		}
	}

	return files, err
}

func (this *GitHubClientImpl) GetFileContentFromGithub(token string, file *File) (string, error) {
	log.Printf("(%v) Checking file - %v\n", file.Filename, file.Sha)
	var content string
	var err error = nil
	content, err = this.getFileContent(token, file.ContentsURL)
	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to access the content of the file for checking - %v", err))
	}
	return content, err
}

func (this *GitHubClientImpl) getFileContent(token string, contentURL string) (string, error) {
	contents := ""

	var err error = nil
	var req *http.Request
	req, err = http.NewRequest("GET", contentURL, nil)
	if err == nil {

		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Accept", "application/vnd.github.v3.raw")
		var resp *http.Response

		log.Printf("Sending HTTP GET to %s", contentURL)

		resp, err = this.httpClient.Do(req)
		if err == nil {

			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				err = errors.New("invalid response from content fetch " + resp.Status)
			} else {

				var bodyBytes []byte
				bodyBytes, err = io.ReadAll(resp.Body)
				if err == nil {

					this.LogHttpPayload(bodyBytes)

					contents = string(bodyBytes)
				}
			}
		}
	}

	return contents, err
}

// Create a 'check run' on github.
func (this *GitHubClientImpl) CreateCheckRun(tokenSupplier TokenSupplier, webhook *Webhook, headSha string) (string, error) {

	var url string = ""

	installationId := webhook.Installation.Id

	var err error = nil
	var token string
	token, err = tokenSupplier.GetToken(installationId)
	if err == nil {

		checkRun := CheckRun{
			Name:    "copyright",
			HeadSha: &headSha,
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
			checkRunsUrl := webhook.Repository.RepositoryURL + "/check-runs"
			req, err = http.NewRequest("POST", checkRunsUrl, bytes.NewReader(checkRunBytes))
			if err == nil {

				req.Header.Add("Authorization", "Bearer "+token)
				req.Header.Add("Accept", "application/vnd.github.v3+json")
				req.Header.Add("Content-Type", "application/vnd.github.v3+json")

				log.Printf("Sending HTTP POST to %s", checkRunsUrl)

				var resp *http.Response
				resp, err = this.httpClient.Do(req)
				if err == nil {

					defer resp.Body.Close()

					if resp.StatusCode != 201 {
						err = errors.New(fmt.Sprintf("Got a non-201 status code from a POST github. status code=%d", resp.StatusCode))
					} else {
						var bodyBytes []byte
						bodyBytes, err = io.ReadAll(resp.Body)
						if err == nil {

							this.LogHttpPayload(bodyBytes)

							var response CheckRun

							err = json.Unmarshal(bodyBytes, &response)
							if err == nil {
								url = *response.Url
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
func (this *GitHubClientImpl) UpdateCheckRun(
	tokenSupplier TokenSupplier,
	webhook *Webhook,
	checkRunURL string,
	checkErrors []checkTypes.CheckError,
	fatalError string,
) error {

	var err error = nil
	var token string

	token, err = tokenSupplier.GetToken(webhook.Installation.Id)
	if err == nil {

		checkRun := CheckRun{
			Name:   "copyright",
			Status: "completed",
			Output: CheckRunOutput{
				Title:   "Galasa copyright check",
				Summary: "Checks for updated copyright years and licence text",
			},
		}

		conclusion := "success"

		if fatalError != "" {
			conclusion = "failure"
			checkRun.Output.Summary = fatalError
		} else if len(checkErrors) > 0 {
			conclusion = "failure"
			annotations := make([]CheckRunAnnotation, 0)

			for _, checkError := range checkErrors {
				annotation := CheckRunAnnotation{
					Path:      checkError.Path,
					Message:   checkError.Message,
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
			req, err = http.NewRequest("PATCH", checkRunURL, bytes.NewReader(checkRunBytes))
			if err == nil {

				req.Header.Add("Authorization", "Bearer "+token)
				req.Header.Add("Accept", "application/vnd.github.v3+json")
				req.Header.Add("Content-Type", "application/vnd.github.v3+json")

				var resp *http.Response
				log.Printf("Sending HTTP PATCH to %s", checkRunURL)
				resp, err = this.httpClient.Do(req)
				if err == nil {

					defer resp.Body.Close()

					var bodyBytes []byte
					bodyBytes, err = io.ReadAll(resp.Body)
					if err == nil {

						data := string(bodyBytes)

						if resp.StatusCode != 200 {
							log.Printf("Fatal error - %v\n", data)
							err = errors.New(fmt.Sprintf("Non-200 status returned from github. %d", resp.StatusCode))
						} else {
							this.LogHttpPayload(bodyBytes)
						}
					}
				}
			}

		}
	}

	if err != nil {
		log.Printf("Fatal error - %v\n", err)
	}

	return err
}
