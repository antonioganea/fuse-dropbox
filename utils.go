/*
	Different utility functions used throughout the project.
*/

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Utility functions for ConstructTreeFromDrpPaths:

// Returns the string after the last '/'.
// E.g: /preda/raluca/antonio -> antonio
// 		/preda/raluca -> raluca
func lastFolderFromPath(path string) string {
	slices := strings.Split(path, "/")
	return slices[len(slices)-1]
}

// Returns the string before the last '/'.
// E.g: /preda/raluca/antonio -> /preda/raluca
// 		/preda/raluca -> /preda
func firstPartFromPath(path string) string {
	return path[:strings.LastIndex(path, "/")]
}

// Utility functions for tokenManager.go :

// Retrieves the AccessToken from its file.
func readToken(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Stores the AccessToken in its file.
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
