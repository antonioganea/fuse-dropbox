package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"golang.org/x/net/context"
)

const (
	configFileName = "AccessToken"
)

var config dropbox.Config

func validatePath(p string) (path string, err error) {
	path = p

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	path = strings.TrimSuffix(path, "/")

	return
}

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

	// Replace here
	var appKey = ""
	var appSecret = ""

	if err != nil {
		conf := oauth2.Config{ // maybe a & reference here
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

func copyOperation() error {
	// Here we do some basic operation : copying file.txt into newfile.txt
	dbx := files.New(config)

	relocArg := files.NewRelocationArg("/file.txt", "/newfile.txt")

	if _, err := dbx.CopyV2(relocArg); err != nil {
		return err
	}
	return nil
}

func downloadOp() {
	dbx := files.New(config)

	downloadArg := files.NewDownloadArg("/file.txt")

	meta, content, err := dbx.Download(downloadArg)
	if err != nil {
		return
	}

	// Here we'd need a better file reading mechanic ( so we know for sure we've read all )
	b1 := make([]byte, meta.Size)
	n1, err := content.Read(b1)
	/////////////////////////

	fmt.Println(string(b1[:n1]))
}

func main() {
	initDbx()
	//copyOperation()
	downloadOp()
}
