package main

import (
	// "encoding/json"
	// "io/ioutil"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Galasa copyright checks")

	http.HandleFunc("/githubapp/copyright/event_handler", eventHandler)
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
