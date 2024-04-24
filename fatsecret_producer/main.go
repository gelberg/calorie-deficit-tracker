package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gelberg/calorie-deficit-tracker/common"
	"github.com/gelberg/oauth1/oauth"

	fat_secret "main.go/pkg" // TODO: fix this import?
)

const (
	RequestIntervalParamName = "FATSECRET_REQUEST_INTERVAL_MS"
	DefaultRequestIntervalMs = 10000
)

func prepareOAuthClient() (oauth.Client, error) {
	oauthClient := oauth.Client{
		TemporaryCredentialRequestURI: "https://www.fatsecret.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://www.fatsecret.com/oauth/authorize",
		TokenRequestURI:               "https://www.fatsecret.com/oauth/access_token",
		SignatureMethod:               oauth.HMACSHA1,
	}

	var credPath = flag.String("client", "client.json", "Path to configuration file containing the client's credentials.")
	b, err := os.ReadFile(*credPath)
	if err != nil {
		return oauth.Client{}, err
	}

	err = json.Unmarshal(b, &oauthClient.Credentials)
	if err != nil {
		return oauth.Client{}, err
	}

	return oauthClient, nil
}

func authorize(oauthClient oauth.Client) (*oauth.Credentials, error) {
	values := url.Values{}
	callbackURL := "oob"
	oauthClient.SignForm(nil, "POST", oauthClient.TemporaryCredentialRequestURI, values, "", callbackURL)
	tempCred, err := oauthClient.RequestTemporaryCredentials(nil, callbackURL, values)
	if err != nil {
		return nil, fmt.Errorf("oauth.Client.RequestTemporaryCredentials: %v", err)
	}

	u := oauthClient.AuthorizationURL(tempCred, nil)

	fmt.Printf("1. Go to %s\n2. Authorize the application\n3. Enter verification code:\n", u)

	var code string
	fmt.Scanln(&code)

	values = url.Values{}
	oauthClient.SignForm(tempCred, "POST", oauthClient.TokenRequestURI, values, code, "")
	tokenCred, _, err := oauthClient.RequestToken(nil, tempCred, code, values)
	if err != nil {
		return nil, fmt.Errorf("oauth.Client.RequestToken: %v", err)
	}

	return tokenCred, nil
}

func prepareTokenCredentials(oauthClient oauth.Client) (*oauth.Credentials, error) {
	var tokenCred *oauth.Credentials
	credPath := flag.String("token", "token.json", "Path to configuration file containing the token's credentials.")
	b, err := os.ReadFile(*credPath)
	if err == nil {
		err = json.Unmarshal(b, &tokenCred)
	}

	if err != nil {
		tokenCred, err = authorize(oauthClient)
		if err != nil {
			return nil, err
		}

		data, _ := json.Marshal(tokenCred)
		err = os.WriteFile(*credPath, data, os.ModeAppend)
		if err != nil {
			log.Printf("Authorization has succeeded, but error occured on token credentials save: %v", err)
		}
	}

	return tokenCred, nil
}

func main() {
	oauthClient, err := prepareOAuthClient()
	if err != nil {
		log.Fatal(err)
	}

	tokenCred, err := prepareTokenCredentials(oauthClient)
	if err != nil {
		log.Fatal(err)
	}

	topic := "consumption"
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

		response, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		var food_entries_resp fat_secret.Response
		json.Unmarshal(response, &food_entries_resp)

		todayConsumption := 0
		for _, food_entry := range food_entries_resp.Food_Entries.Food_Entry {
			i, _ := strconv.Atoi(food_entry.Calories)
			todayConsumption += i
		}

		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err = conn.Write([]byte(fmt.Sprint(todayConsumption)))
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}

		log.Printf("Consumption at %v: %v\n", time.Now(), todayConsumption)

		nextRequest = nextRequest.Add(requestInterval)
		time.Sleep(-time.Since(nextRequest))
	}
}
