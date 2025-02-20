package main

import (
	"crypto/md5"
	"encoding/hex"
)

var alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func bruteForce(targetHash string, maxLength, workerCount, partNumber, partCount int) []string {
	var found []string

	wordsCount := countWordsInAlphabet(alphabet, maxLength)
	start, end := findWordsRangeBounds(wordsCount, partCount, workerCount, partNumber)

	for i := start; i <= end; i++ {
		word := numberToWord(i, alphabet, maxLength)
		hash := md5.Sum([]byte(word))
		hashStr := hex.EncodeToString(hash[:])

		if hashStr == targetHash {
			found = append(found, word)
		}
	}

	return found
}

func findWordsRangeBounds(size, part, n, r int) (int, int) {
	base := size / n
	rem := size % n

	var start int
	if r < rem {
		start = r * (base + 1)
	} else {
		start = r * base
	}

	return start, start + part - 1
}

func countWordsInAlphabet(alphabet string, length int) int {
	n := len(alphabet)
	wordsCount := 0
	for i := 1; i <= length; i++ {
		wordsCount += pow(n, i)
	}
	return wordsCount
}

func pow(x, n int) int {
	if n < 0 {
		return 1 / pow(x, -n)
	}
	if n == 0 {
		return 1
	}
	a := pow(x, n/2)
	if n&1 == 0 {
		return a * a
	}
	return a * a * x
}

func numberToWord(num int, alphabet string, maxLength int) string {
	base := len(alphabet)
	length := 1
	count := base
	for num >= count {
		num -= count
		length++
		count *= base

		if length > maxLength {
			return ""
		}
	}

	word := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		word[i] = alphabet[num%base]
		num /= base
	}

	return string(word)
}
