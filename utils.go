package main

import "strings"

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
