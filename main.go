package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gelberg/oauth1/oauth"
	kafka "github.com/segmentio/kafka-go"
	fat_secret "main.go/pkg" // TODO: fix this import?
)

var oauthClient = oauth.Client{
	TemporaryCredentialRequestURI: "https://www.fatsecret.com/oauth/request_token",
	ResourceOwnerAuthorizationURI: "https://www.fatsecret.com/oauth/authorize",
	TokenRequestURI:               "https://www.fatsecret.com/oauth/access_token",
}

var (
	kafkaEndpoint             = os.Getenv("KAFKA_ENDPOINT")
	fatsecretReqestIntervalMs = os.Getenv("FATSECRET_REQUEST_INTERVAL_MS")
)

func authorize() *oauth.Credentials {
	values := url.Values{}
	oauthClient.SignForm(nil, "POST", oauthClient.TemporaryCredentialRequestURI, values, "", "oob")
	tempCred, err := oauthClient.RequestTemporaryCredentials(nil, "oob", values)
	if err != nil {
		log.Fatal("RequestTemporaryCredentials:", err)
	}

	u := oauthClient.AuthorizationURL(tempCred, nil)

	fmt.Printf("1. Go to %s\n2. Authorize the application\n3. Enter verification code:\n", u)

	var code string
	fmt.Scanln(&code)

	values = url.Values{}
	oauthClient.SignForm(tempCred, "POST", oauthClient.TokenRequestURI, values, code, "")
	tokenCred, _, err := oauthClient.RequestToken(nil, tempCred, code, values)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Authorization has succeeded. Token credentials:\nToken: ", tokenCred.Token, "\nSecret: ", tokenCred.Secret)
	}

	return tokenCred
}

func connectToKafka() (*kafka.Conn, error) {
	topic := "consumption"
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

	err = json.Unmarshal(b, &oauthClient.Credentials)
	if err != nil {
		log.Fatal(err)
	}

	oauthClient.SignatureMethod = oauth.HMACSHA1

	var tokenCred *oauth.Credentials
	credPath = flag.String("token", "token.json", "Path to configuration file containing the token's credentials.")
	b, err = os.ReadFile(*credPath)
	if err == nil {
		err = json.Unmarshal(b, &tokenCred)
	}

	if err != nil {
		tokenCred = authorize()

		data, _ := json.Marshal(tokenCred)
		os.WriteFile(*credPath, data, os.ModeAppend)
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

	requestIntervalMs, err := strconv.Atoi(fatsecretReqestIntervalMs)
	if err != nil {
		log.Fatal(err)
	}

	requestInterval := time.Millisecond * time.Duration(requestIntervalMs)
	nextRequest := time.Now()

	for {
		values := url.Values{
			"method": {"food_entries.get.v2"},
			"format": {"json"},
			// "date":   {"19782"}, // Missing argument means current day
		}

		urlStr := "https://platform.fatsecret.com/rest/server.api"
		oauthClient.SignForm(tokenCred, "GET", urlStr, values, "", "")
		resp, err := oauthClient.Get(nil, tokenCred, urlStr, values)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		response, err := io.ReadAll(resp.Body)

		var food_entries_resp fat_secret.Response
		json.Unmarshal(response, &food_entries_resp)

		calories := 0
		for _, food_entry := range food_entries_resp.Food_Entries.Food_Entry {
			i, _ := strconv.Atoi(food_entry.Calories)
			calories += i
		}

		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte(fmt.Sprint(calories))},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}

		log.Printf("Consumption at %v: %v\n", time.Now(), calories)

		nextRequest = nextRequest.Add(requestInterval)
		time.Sleep(-time.Since(nextRequest))
	}

}