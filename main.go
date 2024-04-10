package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/fitness/v1"
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
	authURL := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", authURL)

	var code string
	fmt.Scanln(&code)

	// Handle the exchange code to initiate a transport.
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}
	client := conf.Client(oauth2.NoContext, tok)

	svc, err := fitness.New(client)
	if err != nil {
		log.Fatalf("Unable to create Fitness service: %v", err)
	}

	for {
		year, month, day := time.Now().Date() // TODO: fix discrepancy between time.Now() on local PC (UTC+4) and Docker (UTC+0)
		todayStart := time.Date(year, month, day, 0, 0, 0, 0, time.FixedZone("GET", 4*60*60))
		tomorrowStart := todayStart.Add(24 * time.Hour)

		aggrReq := fitness.AggregateRequest{
			AggregateBy: []*fitness.AggregateBy{{
				DataTypeName: "com.google.calories.expended",
				DataSourceId: "derived:com.google.calories.expended:com.google.android.gms:merge_calories_expended",
			}},
			BucketByTime: &fitness.BucketByTime{
				DurationMillis: 24 * time.Hour.Milliseconds(),
			},
			StartTimeMillis: todayStart.UnixMilli(),
			EndTimeMillis:   tomorrowStart.UnixMilli(),
		}
		r, e := svc.Users.Dataset.Aggregate("me", &aggrReq).Do()
		if e != nil {
			log.Fatal(e)
		} else {
			fmt.Println(*&r.Bucket[0].Dataset[0].Point[0].Value[0].FpVal)
		}

		time.Sleep(5 * time.Second) // TODO: configurable parameter?
	}
}
