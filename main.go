package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	kafka "github.com/segmentio/kafka-go"
	"google.golang.org/api/fitness/v1"
)

var (
	kafkaEndpoint = os.Getenv("KAFKA_ENDPOINT")
)

func connectToKafka() (*kafka.Conn, error) {
	topic := "expenditure"
	partition := 0
	// Workaround for the following scenario:
	// 1. docker-compose with kafka and this service is started:
	// 1.1. kafka is started
	// 1.2. this service attempts to connect to kafka, gets connection refuse and exits
	// 1.3. kafka initializes its listeners
	connectionAttempts := 3
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaEndpoint, topic, partition)
	for connectionAttempts > 0 && err != nil { // TODO: distinguish our case and others?
		conn, err = kafka.DialLeader(context.Background(), "tcp", kafkaEndpoint, topic, partition)
		connectionAttempts--
		time.Sleep(time.Second)
	}

	return conn, err
}

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

	conn, err := connectToKafka()
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal("failed to close writer:", err)
		}
	}()

	requestIntervalMs, err := strconv.Atoi(googleFitReqestIntervalMs)
	if err != nil {
		log.Fatal(err)
	}

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

		calories := 0
		r, e := svc.Users.Dataset.Aggregate("me", &aggrReq).Do()
		if e != nil {
			log.Fatal(e)
		} else {
			for _, b := range r.Bucket {
				for _, d := range b.Dataset {
					for _, p := range d.Point {
						for _, v := range p.Value {
							calories += int(v.FpVal)
						}
					}
				}
			}

			log.Printf("Consumption at %v: %v\n", time.Now(), calories)
		}

		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte(fmt.Sprint(calories))},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}

		nextRequest = nextRequest.Add(requestInterval)
		time.Sleep(-time.Since(nextRequest))
	}
}
