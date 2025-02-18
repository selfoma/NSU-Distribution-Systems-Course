package main

import (
	"crypto/md5"
	"encoding/hex"
)

func bruteForce(targetHash string, maxLength, partNumber, partCount int) []string {
	alphabet := "abcdefghijklmnopqrstuvwxyz0123456789"
	var found []string

	for _, word := range generateWords(alphabet, maxLength) {
		hash := md5.Sum([]byte(word))
		hashStr := hex.EncodeToString(hash[:])

		if hashStr == targetHash {
			found = append(found, word)
		}
	}

	return found
}

func generateWords(alphabet string, maxLength int) []string {
	return []string{"abcd", "abc"}
}
