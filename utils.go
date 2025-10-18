package main

import (
	"slices"
	"strings"
)

var bannedWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

// string utils
func Cenzo(msg string) string {
	words := strings.Fields(msg)
	for i, word := range words {
		lowerword := strings.ToLower(word)
		if slices.Contains(bannedWords, lowerword) {
			words[i] = "****"
		}
	}
	return strings.Join(words," ")
}
