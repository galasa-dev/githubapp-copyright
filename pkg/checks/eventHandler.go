/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/galasa-dev/githubapp-copyright/pkg/checkTypes"
)

type EventHandler interface {
	HandleEvent(w http.ResponseWriter, r *http.Request)
}

type EventHandlerImpl struct {
	checker       Checker
	tokenSupplier TokenSupplier
}

func NewEventHandlerImpl(checker Checker, tokenSupplier TokenSupplier) (EventHandler, error) {
	var err error = nil
	this := new(EventHandlerImpl)
	this.checker = checker
	this.tokenSupplier = tokenSupplier
	return this, err
}

func (this *EventHandlerImpl) HandleEvent(w http.ResponseWriter, r *http.Request) {

	log.Println("Inbound event")
	var status int = http.StatusOK

	expectedUrlPath := "/githubapp/copyright/event_handler"
	if r.URL.Path != expectedUrlPath {
		log.Printf("Failed: Bad request. Request is not made to expected path %s", expectedUrlPath)
		status = http.StatusNotFound
	} else {

		if r.Method != "POST" {
			log.Printf("Failed: Bad request. Request type is not a POST")
			status = http.StatusMethodNotAllowed
		} else {

			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				log.Printf("Failed: Bad request. Content type is not application/json.")
				status = http.StatusUnsupportedMediaType
			} else {

				var jsonBytes []byte
				var err error = nil
				jsonBytes, err = io.ReadAll(r.Body)
				if err != nil {
					log.Printf("Failed to read the request body. Ignoring. reason: %v", err)
					status = http.StatusInternalServerError
				} else {

					var webhook Webhook
					err = json.Unmarshal(jsonBytes, &webhook)
					if err != nil {
						log.Printf("Parse webhook failed - %v\n", err)
						status = http.StatusInternalServerError
					} else {

						log.Printf("    Received action %v\n", webhook.Action)

						if webhook.CheckSuite != nil {
							go this.performCheckSuite(&webhook)
						} else if webhook.CheckRun != nil {
							go this.performCheckRun(&webhook)
						} else if webhook.Action == "opened" || webhook.Action == "synchronize" {
							go this.performPullRequest(&webhook)
						}
					}
				}
			}

		}
	}

	w.WriteHeader(status)
}

func (this *EventHandlerImpl) performCheckSuite(webhook *Webhook) error {

	var err error = nil

	if webhook.Action != "requested" {
		err = errors.New("Error: Webhook action is not 'requested' so not performing suite check.")
	} else {

		log.Printf("Performing check suite tests on (%v) - repository %v\n", webhook.CheckSuite.Id, webhook.Repository.RepositoryURL)
		var checkRunURL string
		if len(*webhook.CheckSuite.PullRequests) > 0 {
			// We have pull requests so will use that to obtain a list of files to check

			checkRunURL, err = CreateCheckRun(this.tokenSupplier, webhook, webhook.CheckSuite.HeadSha)

			if err == nil {
				pullRequests := webhook.CheckSuite.PullRequests
				errors := this.performPullRequestChecks(webhook, webhook.CheckSuite.Id, checkRunURL, pullRequests)

				if len(*errors) > 0 {
					log.Printf("(%v) Errors found with check suite", webhook.CheckSuite.Id)
				}
			}
		} else if webhook.CheckSuite.Before != nil && webhook.CheckSuite.After != nil {
			checkRunURL, err = CreateCheckRun(this.tokenSupplier, webhook, webhook.CheckSuite.HeadSha)
			if err == nil {

				var checkErrors []checkTypes.CheckError
				checkErrors, err = this.performBeforeAfterChecks(webhook, webhook.CheckSuite.Id, checkRunURL, *webhook.CheckSuite.Before, *webhook.CheckSuite.After)
				if len(checkErrors) > 0 {
					log.Printf("(%v) Errors found with check suite", webhook.CheckSuite.Id)
				}
			}
		} else {
			log.Println("Unrecognised payload for check suite")
		}
	}

	if err != nil {
		log.Printf("Error: Failed to check suite. Reason: %s\n", err.Error())
	}

	return err
}

func (this *EventHandlerImpl) performCheckRun(webhook *Webhook) error {

	var err error = nil

	if webhook.Action != "rerequested" {
		err = errors.New("Failed to perform check run because webhook action is not rerequested.")
	} else {

		log.Printf("Performing check run tests on (%v) - repository %v\n", webhook.CheckRun.Id, webhook.Repository.RepositoryURL)

		var checkRunURL string
		if len(*webhook.CheckRun.CheckSuite.PullRequests) > 0 {
			checkRunURL, err = CreateCheckRun(this.tokenSupplier, webhook, webhook.CheckRun.HeadSha)

			if err == nil {
				// We have pull requests so will use that to obtain a list of files to check
				pullRequests := webhook.CheckRun.CheckSuite.PullRequests
				errors := this.performPullRequestChecks(webhook, webhook.CheckRun.Id, checkRunURL, pullRequests)

				if len(*errors) > 0 {
					log.Printf("(%v) Errors found with check run", webhook.CheckRun.Id)
				}
			}
		} else if webhook.CheckRun.CheckSuite.Before != nil && webhook.CheckRun.CheckSuite.After != nil {
			checkRunURL, err = CreateCheckRun(this.tokenSupplier, webhook, webhook.CheckRun.HeadSha)
			if err == nil {
				var checkErrors []checkTypes.CheckError
				checkErrors, err = this.performBeforeAfterChecks(
					webhook, webhook.CheckRun.Id, checkRunURL,
					*webhook.CheckRun.CheckSuite.Before, *webhook.CheckRun.CheckSuite.After)
				if len(checkErrors) > 0 {
					log.Printf("(%v) Errors found with check run", webhook.CheckRun.Id)
				}
			}
		} else {
			log.Println("Unrecognised payload for check run")
		}
	}

	if err != nil {
		log.Printf("Error: Failed to check run. Reason: %s\n", err.Error())
	}

	return err
}

func (this *EventHandlerImpl) performPullRequest(webhook *Webhook) error {
	var err error = nil

	if webhook.PullRequest == nil {
		err = errors.New("Cannot process a null pull request ")
	} else {
		if webhook.PullRequest.Head.Sha == "" {
			err = errors.New("Cannot process a pull request with an empty Sha")
		} else {
			log.Printf("Performing pull request open tests on (%v) - repository %v\n", webhook.PullRequest.Number, webhook.Repository.RepositoryURL)

			if webhook.Action == "synchronize" {
				if webhook.PullRequest.Head.Repo.Id == webhook.PullRequest.Base.Repo.Id {
					err = errors.New(fmt.Sprintf("(%v) ignoring pr sync for same repo prs, as rerequest should be issued", webhook.PullRequest.Number))
				}
			}

			if err == nil {

				var checkRunURL string
				checkRunURL, err = CreateCheckRun(this.tokenSupplier, webhook, webhook.PullRequest.Head.Sha)

				if err == nil {
					pullRequests := make([]WebhookPullRequest, 0)
					pullRequests = append(pullRequests, *webhook.PullRequest)

					checkErrors := this.performPullRequestChecks(webhook, webhook.PullRequest.Number, checkRunURL, &pullRequests)

					if len(*checkErrors) > 0 {
						err = errors.New(fmt.Sprintf("(%v) Errors found with pull request open", webhook.PullRequest.Number))
					}
				}
			}
		}
	}

	if err != nil {
		log.Printf("Error: Failed to check run. Reason: %s\n", err.Error())
	}

	return err
}

func (this *EventHandlerImpl) performPullRequestChecks(webhook *Webhook, checkId int, checkRunURL string, pullRequests *[]WebhookPullRequest) *[]checkTypes.CheckError {

	checkErrors := make([]checkTypes.CheckError, 0)

	for _, pr := range *pullRequests {
		var err error
		var newCheckErrors *[]checkTypes.CheckError
		newCheckErrors, err = this.checker.CheckPullRequest(webhook, checkId, pr.Url)
		if err != nil {
			log.Printf("(%v) Fatal error - %v", checkId, err)
			fatalError := fmt.Sprintf("Fatal error - %v", err)
			UpdateCheckRun(this.tokenSupplier, webhook, checkRunURL, &checkErrors, fatalError)
		}
		if newCheckErrors != nil {
			for _, newError := range *newCheckErrors {
				checkErrors = append(checkErrors, newError)
			}
		}
	}

	UpdateCheckRun(this.tokenSupplier, webhook, checkRunURL, &checkErrors, "")

	return &checkErrors
}

func (this *EventHandlerImpl) performBeforeAfterChecks(webhook *Webhook, checkId int, checkRunURL string, before string, after string) ([]checkTypes.CheckError, error) {
	log.Printf("(%v) Checking commit '%v'->'%v'", checkId, before, after)

	var checkErrors []checkTypes.CheckError = nil
	var err error = nil
	var token string

	token, err = this.tokenSupplier.GetToken(webhook.Installation.Id)

	if err == nil {
		var filesURL string
		filesURL, err = this.calculateFilesUrl(webhook, checkId, checkRunURL, before, after)

		var allFiles []File
		client := &http.Client{}

		allFiles, err = getFilesChanged(client, token, filesURL)

		for _, file := range allFiles {
			var newCheckError *checkTypes.CheckError
			newCheckError = this.checker.CheckFile(webhook, checkId, &token, client, &file)

			if newCheckError != nil {
				log.Printf("(%v) Found problem with file %v - %v", checkId, file.Filename, newCheckError.Message)
				checkErrors = append(checkErrors, *newCheckError)
			}

			// Continue to check the next file also.
		}

		if err == nil {
			UpdateCheckRun(this.tokenSupplier, webhook, checkRunURL, &checkErrors, "")
		}
	}

	return checkErrors, err
}

func (this *EventHandlerImpl) setAdhocError(webhook *Webhook, checkId int, checkRunURL string, message string) {
	log.Printf("(%v) %v", checkId, message)

	UpdateCheckRun(this.tokenSupplier, webhook, checkRunURL, nil, message)
}

func (this *EventHandlerImpl) calculateFilesUrl(webhook *Webhook, checkId int, checkRunURL string, before string, after string) (string, error) {
	var err error = nil
	filesURL := ""
	if before != "0000000000000000000000000000000000000000" {
		filesURL = webhook.Repository.CompareURL
		if filesURL == "" {
			this.setAdhocError(webhook, checkId, checkRunURL, "request is missing compare_url")
		} else {

			// Retrieve the list of files in a compare
			filesURL = strings.Replace(filesURL, "{base}", before, 1)
			filesURL = strings.Replace(filesURL, "{head}", after, 1)
		}
	} else {
		filesURL = webhook.Repository.CommitsURL
		if filesURL == "" {
			this.setAdhocError(webhook, checkId, checkRunURL, "request is missing commits_url")
		} else {
			filesURL = strings.Replace(filesURL, "{/sha}", "/"+after, 1)
		}
	}
	return filesURL, err
}
