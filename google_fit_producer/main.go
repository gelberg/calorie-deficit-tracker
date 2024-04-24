package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gelberg/calorie-deficit-tracker/common"
	"google.golang.org/api/fitness/v1"
)

const (
	RequestIntervalParamName = "GOOGLE_FIT_REQUEST_INTERVAL_MS"
	DefaultRequestIntervalMs = 10000
)

func prepareConfig() (*oauth2.Config, error) {
	var credPath = flag.String("client", "client.json", "Path to configuration file containing the client's credentials.")
	b, err := os.ReadFile(*credPath)
	if err != nil {
		return nil, err
	}

	var conf *oauth2.Config
	json.Unmarshal(b, &conf)

	conf.Scopes = []string{"https://www.googleapis.com/auth/fitness.activity.read"}
	conf.Endpoint = google.Endpoint

	return conf, nil
}

func authorize(conf *oauth2.Config) (*http.Client, error) {
	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	authURL := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", authURL)

	var code string
	fmt.Scanln(&code)

	// Handle the exchange code to initiate a transport.
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	client := conf.Client(oauth2.NoContext, tok)
	return client, nil
}

func main() {
	conf, err := prepareConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := authorize(conf)
	if err != nil {
		log.Fatal(err)
	}

	svc, err := fitness.New(client)
	if err != nil {
		log.Fatalf("Unable to create Fitness service: %v", err)
	}

	topic := "expenditure"
	conn, err := common.ConnectToKafka(topic)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal("failed to close writer:", err)
		}
	}()

	requestIntervalMs := common.GetEnvInt(RequestIntervalParamName, DefaultRequestIntervalMs)
	requestInterval := time.Millisecond * time.Duration(requestIntervalMs)
	nextRequest := time.Now()

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

		todayExpenditure := 0
		resp, err := svc.Users.Dataset.Aggregate("me", &aggrReq).Do()
		if err != nil {
			log.Fatal(err)
		} else {
			for _, b := range resp.Bucket {
				for _, d := range b.Dataset {
					for _, p := range d.Point {
						for _, v := range p.Value {
							todayExpenditure += int(v.FpVal)
						}
					}
				}
			}

			log.Printf("Expenditure at %v: %v\n", time.Now(), todayExpenditure)
		}

		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err = conn.Write([]byte(fmt.Sprint(todayExpenditure)))
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}

		nextRequest = nextRequest.Add(requestInterval)
		time.Sleep(-time.Since(nextRequest))
	}
}
