package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gomodule/oauth1/oauth"
)

func main() {
	client := oauth.Client{}
	client.Credentials.Token = "***REMOVED***"
	client.Credentials.Secret = "***REMOVED***"
	client.SignatureMethod = oauth.HMACSHA1

	values := url.Values{
		"method": {"food_entries.get_month"},
		"format": {"json"},
		"date":   {"19782"},
	}
	u, err := url.Parse("https://platform.fatsecret.com/rest/server.api")
	if err != nil {
		log.Fatal(err)
	}

	creds := oauth.Credentials{Token: "***REMOVED***", Secret: "***REMOVED***"}

	fmt.Println("values: ", values)
	client.SignForm(&creds, "food_entries.get_month", u.String(), values)
	fmt.Println("values: ", values)

	fmt.Println("Header: ", client.Header)
	fmt.Println("Credentials: ", client.Credentials)
	header := http.Header{}
	err = client.SetAuthorizationHeader(header, &creds, "food_entries.get_month", u, values)
	if err != nil {
		log.Fatal(err)
		return
	}
	client.Header = header
	fmt.Println("Header: ", client.Header)
	fmt.Println("Credentials: ", client.Credentials)

	resp, err := client.Get(nil, &creds, u.String(), values)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
