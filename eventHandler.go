/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func eventHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Inbound event")

	if r.URL.Path != "/githubapp/copyright/event_handler" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	jsonBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var webhook Webhook
	err = json.Unmarshal(jsonBytes, &webhook)
	if err != nil {
		log.Fatalf("Parse webhook failed - %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("    Received action %v\n", webhook.Action)

	if webhook.CheckSuite != nil {
		go performCheckSuite(&webhook)
	} else if webhook.CheckRun != nil {
		go performCheckRun(&webhook)
	} else if webhook.Action == "opened" || webhook.Action == "synchronize"  {
		go performPullRequest(&webhook)
	}


	w.WriteHeader(http.StatusOK)
}


func performCheckSuite(webhook *Webhook) {

	if webhook.Action != "requested" {
		return
	}

	log.Printf("Performing check suite tests on (%v) - repository %v\n", webhook.CheckSuite.Id, webhook.Repository.RepositoryURL)

	if len(*webhook.CheckSuite.PullRequests) > 0 {
		// We have pull requests so will use that to obtain a list of files to check
		checkRunURL := createCheckRun(webhook, &webhook.CheckSuite.HeadSha)

		pullRequests := webhook.CheckSuite.PullRequests
		errors := performPullRequestChecks(webhook, webhook.CheckSuite.Id, checkRunURL, pullRequests)

		if len(*errors) > 0 {
			log.Printf("(%v) Errors found with check suite", webhook.CheckSuite.Id)
		}
	} else if webhook.CheckSuite.Before != nil && webhook.CheckSuite.After != nil {
		checkRunURL := createCheckRun(webhook, &webhook.CheckSuite.HeadSha)

		errors := performBeforeAfterChecks(webhook, webhook.CheckSuite.Id, checkRunURL, webhook.CheckSuite.Before, webhook.CheckSuite.After)
		if len(*errors) > 0 {
			log.Printf("(%v) Errors found with check suite", webhook.CheckSuite.Id)
		}
	} else {
		log.Println("Unrecognised payload for check suite")
	}
}


func performCheckRun(webhook *Webhook) {
	
	if webhook.Action != "rerequested" {
		return
	}

	log.Printf("Performing check run tests on (%v) - repository %v\n", webhook.CheckRun.Id, webhook.Repository.RepositoryURL)

	if len(*webhook.CheckRun.CheckSuite.PullRequests) > 0 {
		checkRunURL := createCheckRun(webhook, &webhook.CheckRun.HeadSha)
		// We have pull requests so will use that to obtain a list of files to check
		pullRequests := webhook.CheckRun.CheckSuite.PullRequests
		errors := performPullRequestChecks(webhook, webhook.CheckRun.Id, checkRunURL, pullRequests)

		if len(*errors) > 0 {
			log.Printf("(%v) Errors found with check run", webhook.CheckRun.Id)
		}
	} else if webhook.CheckRun.CheckSuite.Before != nil && webhook.CheckRun.CheckSuite.After != nil {
		checkRunURL := createCheckRun(webhook, &webhook.CheckRun.HeadSha)
		errors := performBeforeAfterChecks(webhook, webhook.CheckRun.Id, checkRunURL, webhook.CheckRun.CheckSuite.Before, webhook.CheckRun.CheckSuite.After)
		if len(*errors) > 0 {
			log.Printf("(%v) Errors found with check run", webhook.CheckRun.Id)
		}
	} else {
		log.Println("Unrecognised payload for check run")
	}
}

func performPullRequest(webhook *Webhook) {
	if webhook.PullRequest == nil || webhook.PullRequest.Head.Sha == "" {
		return
	}

	log.Printf("Performing pull request open tests on (%v) - repository %v\n", webhook.PullRequest.Number, webhook.Repository.RepositoryURL)

	if webhook.Action == "synchronize" {
		if webhook.PullRequest.Head.Repo.Id == webhook.PullRequest.Base.Repo.Id {
			log.Printf("(%v) ignoring pr sync for same repo prs, as rerequest should be issued", webhook.PullRequest.Number)
			return
		}
	}


	checkRunURL := createCheckRun(webhook, &webhook.PullRequest.Head.Sha)

	pullRequests := make([]WebhookPullRequest, 0)
	pullRequests = append(pullRequests, *webhook.PullRequest)

	errors := performPullRequestChecks(webhook, webhook.PullRequest.Number, checkRunURL, &pullRequests)

	if len(*errors) > 0 {
		log.Printf("(%v) Errors found with pull request open", webhook.PullRequest.Number)
	}

}


func performPullRequestChecks(webhook *Webhook, checkId int, checkRunURL *string, pullRequests *[]WebhookPullRequest) *[]CheckError {

	errors := make([]CheckError, 0)

	for _, pr := range *pullRequests {
		newErrors, err := checkPullRequest(webhook, checkId, pr.Url)
		if err != nil {
			log.Fatalf("(%v) Fatal error - %v", checkId, err)
			fatalError := fmt.Sprintf("Fatal error - %v", err)
			updateCheckRun(webhook, checkRunURL, &errors, &fatalError)
		}
		if newErrors != nil {
			for _, newError := range *newErrors {
				errors = append(errors, newError)
			}
		}
	}

	updateCheckRun(webhook, checkRunURL, &errors, nil)

	return &errors
}



func performBeforeAfterChecks(webhook *Webhook, checkId int, checkRunURL *string, before *string, after *string) *[]CheckError {
	log.Printf("(%v) Checking commit '%v'->'%v'", checkId, *before, *after)

	token := getToken(webhook.Installation.Id)

	errors := make([]CheckError, 0)

	filesURL := ""
	if *before != "0000000000000000000000000000000000000000" {
		filesURL = webhook.Repository.CompareURL
		if filesURL == "" {
			setAdhocError(webhook, checkId, checkRunURL, "request is missing compare_url")
			return nil
		}

		// Retrieve the list of files in a compare
		filesURL = strings.Replace(filesURL, "{base}", *before, 1)
		filesURL = strings.Replace(filesURL, "{head}", *after, 1)
	} else {
		filesURL = webhook.Repository.CommitsURL
		if filesURL == "" {
			setAdhocError(webhook, checkId, checkRunURL, "request is missing commits_url")
			return nil
		}

		filesURL = strings.Replace(filesURL, "{/sha}", "/" + *after, 1)
	}

	client := &http.Client{}
	for page := 1; ; page++ {

		pageUrl := fmt.Sprintf("%v?page=%v", filesURL, page)
		req, err := http.NewRequest("GET", pageUrl, nil)
		if err != nil {
			fatalError := fmt.Sprintf("Fatal error - %v", err)
			setAdhocError(webhook, checkId, checkRunURL, fatalError)
			return nil
		}
		req.Header.Add("Authorization", "Bearer "+ token)
		req.Header.Add("Accept", "application/vnd.github.v3.raw")
		resp, err := client.Do(req)
		if err != nil {
			fatalError := fmt.Sprintf("Fatal error - %v", err)
			setAdhocError(webhook, checkId, checkRunURL, fatalError)
			return nil
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			fatalError := fmt.Sprintf("invalid response from content fetch - %v", resp.Status)
			setAdhocError(webhook, checkId, checkRunURL, fatalError)
			return nil
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fatalError := fmt.Sprintf("Fatal error - %v", err)
			setAdhocError(webhook, checkId, checkRunURL, fatalError)
			return nil
		}

		var files Files

		err = json.Unmarshal(bodyBytes, &files)
		if err != nil {
			fatalError := fmt.Sprintf("Fatal error - %v", err)
			setAdhocError(webhook, checkId, checkRunURL, fatalError)
			return nil
		}

		// If there are no files...
		if files.Files == nil || len(*files.Files) < 1 {
			break
		}

		for _, file := range *files.Files {
			newError := checkFile(webhook, checkId, &token, client, &file)
			if newError != nil {
				log.Printf("(%v) Found problem with file %v - %v", checkId, file.Filename, newError.Message)
				errors = append(errors, *newError)
			}
		}
	}

	updateCheckRun(webhook, checkRunURL, &errors, nil)

	return &errors
}



func setAdhocError(webhook *Webhook, checkId int, checkRunURL *string, message string) {

	log.Fatalf("(%v) %v", checkId, message)
	updateCheckRun(webhook, checkRunURL, nil, &message)

}