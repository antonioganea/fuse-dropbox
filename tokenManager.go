package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"golang.org/x/oauth2"
)

const (
	configFileName = "AccessToken"
)

var config dropbox.Config

func readToken(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeToken(filePath string, token string) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Doesn't exist; lets create it
		err = os.MkdirAll(filepath.Dir(filePath), 0700)
		if err != nil {
			return
		}
	}
	b := []byte(token)
	if err := ioutil.WriteFile(filePath, b, 0600); err != nil {
		return
	}
}

func initDbx() (err error) {

	memorizedToken, err := readToken(configFileName)

	if err != nil {
		conf := oauth2.Config{
			ClientID:     appKey,
			ClientSecret: appSecret,
			Endpoint:     dropbox.OAuthEndpoint(""),
		}

		fmt.Printf("1. Go to %v\n", conf.AuthCodeURL("state"))
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n")
		fmt.Printf("Enter the authorization code here: ")

		var code string
		if _, err = fmt.Scan(&code); err != nil {
			return
		}
		var token *oauth2.Token
		ctx := context.Background()
		token, err = conf.Exchange(ctx, code)
		if err != nil {
			return
		}
		memorizedToken = token.AccessToken
	}

	writeToken(configFileName, memorizedToken)

	config = dropbox.Config{
		Token:    memorizedToken,
		LogLevel: dropbox.LogOff, // if needed, set the desired logging level. Default is off
	}

	fmt.Print("Dropbox Config'd!\n")

	return
}
