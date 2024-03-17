package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/gomodule/oauth1/oauth"
)

var oauthClient = oauth.Client{
	TemporaryCredentialRequestURI: "https://www.fatsecret.com/oauth/request_token",
	ResourceOwnerAuthorizationURI: "https://www.fatsecret.com/oauth/authorize",
	TokenRequestURI:               "https://www.fatsecret.com/oauth/access_token",
}

func main() {
	oauthClient.Credentials.Token = "***REMOVED***"
	oauthClient.Credentials.Secret = "***REMOVED***"
	oauthClient.SignatureMethod = oauth.HMACSHA1

	values := url.Values{
		// "oauth_callback": {"oob"},
	}
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
	}

	values = url.Values{
		"method": {"food_entries.get_month"},
		"format": {"json"},
		"date":   {"19782"},
	}
	urlStr := "https://platform.fatsecret.com/rest/server.api"
	oauthClient.SignForm(tokenCred, "GET", urlStr, values, "", "")
	resp, err := oauthClient.Get(nil, tokenCred, urlStr, values)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatal(err)
	}
}
