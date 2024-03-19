package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	var credPath = flag.String("client", "client.json", "Path to configuration file containing the client's credentials.")
	b, err := os.ReadFile(*credPath)
	if err != nil {
		log.Fatal(err)
	}

	var conf *oauth2.Config
	json.Unmarshal(b, &conf)

	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	conf.Scopes = []string{"https://www.googleapis.com/auth/fitness.activity.read"}
	conf.Endpoint = google.Endpoint
	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	authURL := conf.AuthCodeURL("state")
	fmt.Printf("Visit the URL for the auth dialog: %v\n", authURL)

	var code string
	fmt.Scanln(&code)

	// Handle the exchange code to initiate a transport.
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}
	client := conf.Client(oauth2.NoContext, tok)

	reqBody := `
	{
		"aggregateBy": [{
		  "dataTypeName": "calories",
		  "dataSourceId": "derived:com.google.calories.expended:com.google.android.gms:merge_calories_expended"
		}],
		"bucketByTime": { "durationMillis": 86400000 },
		"startTimeMillis": 1710705599000,
		"endTimeMillis": 1710791999000
	}`
	reqBodyReader := strings.NewReader(reqBody)

	resp, err := client.Post("https://www.googleapis.com/fitness/v1/users/me/dataset:aggregate", "application/json", reqBodyReader)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatal(err)
	}
}
