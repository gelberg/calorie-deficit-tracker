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
	oauthClient.SignForm(nil, "POST", oauthClient.TemporaryCredentialRequestURI, values, "oob")
	tempCred, err := oauthClient.RequestTemporaryCredentials(nil, "oob", values)
	if err != nil {
		log.Fatal("RequestTemporaryCredentials:", err)
	}

	u := oauthClient.AuthorizationURL(tempCred, nil)

	fmt.Printf("1. Go to %s\n2. Authorize the application\n3. Enter verification code:\n", u)

	var code string
	fmt.Scanln(&code)

	tokenCred, _, err := oauthClient.RequestToken(nil, tempCred, code)
	if err != nil {
		log.Fatal(err)
	}

	oauthClient.SignForm(nil, "GET", "https://platform.fatsecret.com/rest/server.api", values, "")
	resp, err := oauthClient.Get(nil, tokenCred,
		"https://platform.fatsecret.com/rest/server.api?method=food_entries.get_month&format=json&date=19782", values)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatal(err)
	}
}
