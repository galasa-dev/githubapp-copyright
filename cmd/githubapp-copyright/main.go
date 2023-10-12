/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	// "encoding/json"
	// "io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/galasa-dev/githubapp-copyright/pkg/checks"
	embedded "github.com/galasa-dev/githubapp-copyright/pkg/embedded"
)

func main() {

	log.Printf("Copyright Checker version %s\n", embedded.GetVersion())
	log.Println("Starting Galasa copyright checks...")

	var err error = nil

	var console checks.Console
	console, err = checks.NewConsole()
	if err == nil {

		var parser checks.CommandLineArgParser
		parser, err = checks.NewCommandLineArgParserImpl(os.Args, console)

		if err == nil {
			var parsedValues *checks.FieldValuesParsed
			parsedValues, err = parser.Parse()
			if err == nil {

				gitHubClient := checks.NewGitHubClient(parsedValues.IsDebugEnabled)

				var tokenSupplier checks.TokenSupplier
				tokenSupplier, err = checks.NewTokenSupplier(gitHubClient, parsedValues.GithubAuthKeyFilePath)
				if err == nil {

					var checker checks.Checker
					checker, err = checks.NewChecker(gitHubClient)
					if err == nil {

						var eventHandler checks.EventHandler
						eventHandler, err = checks.NewEventHandlerImpl(gitHubClient, checker, tokenSupplier)
						if err == nil {

							http.HandleFunc("/githubapp/copyright/event_handler", eventHandler.HandleEvent)
							log.Printf("Listening for http traffic on port 3000...\n")
							err = http.ListenAndServe(":3000", nil)
						}
					}
				}
			}
		}
	}

	if err != nil {
		log.Printf("Failure: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
