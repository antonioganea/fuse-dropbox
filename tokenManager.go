/*
	This file implements routines that make sure you have
	your dropbox account set in place.

	You can either use the appKey and appSecret from the
	developers' console on dropbox.com

	Or you can copy paste the longer access token into a
	fille called AccessToken that should be inside this
	directory, next to the Go scripts.

	The latter one is easier.

	Essentially, if you use the appKey+appSecret, you have
	to paste them in the credentials.go script and run
	the program, you will have some instructions in the
	terminal in order to approve the usage of the app.
	( Dropbox permissions ).

	But in the end it all boils down to getting the
	AccessToken file filled with the required token.
*/

package main

import (
	"context"
	"fmt"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"golang.org/x/oauth2"
)

const (
	// This file stores the AccessToken between runs
	configFileName = "AccessToken"
)

// Global variable that holds dropbox auth configurations
// It is used each time you create a dropbox context,
// and that is required for any dropbox API call.
var config dropbox.Config

// This is a function that is executed even before the
// filesystem is mounted. It checks that you have the
// dropbox account correctly set up.
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
