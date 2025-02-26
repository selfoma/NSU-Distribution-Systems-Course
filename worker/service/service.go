package service

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/selfoma/crackhash/worker/broker"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
)

type WorkerResponse struct {
	RequestId  string   `xml:"requestId"`
	Words      []string `xml:"words"`
	PartNumber int      `xml:"partNumber"`
}

func BruteForce(task broker.WorkerTask) {
	var found []string

	targetHash, length := task.Hash, task.MaxLength

	wordsCount := countWordsInAlphabet(alphabet, length)
	start, end := findWordsRangeBounds(wordsCount, task.PartCount, task.WorkerCount, task.PartNumber)

	for i := start; i <= end; i++ {
		word := numberToWord(i, alphabet, length)
		hash := md5.Sum([]byte(word))
		hashStr := hex.EncodeToString(hash[:])

		if hashStr == targetHash {
			found = append(found, word)
		}
	}

	resp := WorkerResponse{
		RequestId:  task.RequestId,
		Words:      found,
		PartNumber: task.PartNumber,
	}

	broker.PublishResponse(resp)
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
