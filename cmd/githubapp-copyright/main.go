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

	"github.com/galasa-dev/githubapp-copyright/pkg/checks"
)

func main() {
	log.Println("Starting Galasa copyright checks")

	http.HandleFunc("/githubapp/copyright/event_handler", checks.EventHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))

	// jsonBytes, err := ioutil.ReadFile("opened.json")
	// if err != nil {
	// 	panic(err)
	// }

	// var webhook Webhook

	// err = json.Unmarshal(jsonBytes, &webhook)
	// if err != nil {
	// 	panic(err)
	// }

	// performPullRequest(&webhook)
}
