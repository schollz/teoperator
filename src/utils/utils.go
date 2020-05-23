package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	mrand "math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// GetStringInBetween returns empty string if no start or end string found
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}

// TempFileName generates a temporary filename for use in testing or whatever
func TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

// ConvertToSeconds converts a string lik 00:00:11.35 into seconds (11.35)
func ConvertToSeconds(s string) (seconds float64, err error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	multipliers := []float64{60 * 60, 60, 1}
	if len(parts) == 2 {
		multipliers = []float64{60, 1, 1}
	} else if len(parts) == 1 {
		multipliers = []float64{1, 1, 1}
	}
	for i, part := range parts {
		var partf float64
		partf, err = strconv.ParseFloat(part, 64)
		if err != nil {
			return
		}
		seconds += partf * multipliers[i]
	}
	return
}

// SecondsToString seconds like 80 to a string like 00:01:20.00
func SecondsToString(seconds float64) string {
	hours := math.Floor(seconds / 3600)
	seconds = seconds - hours*3600

	minutes := math.Floor(seconds / 60)
	seconds = seconds - minutes*60

	s := fmt.Sprintf("%02d:%02d:%02.4f", int(hours), int(minutes), seconds)
	if seconds < 10 {
		s = fmt.Sprintf("%02d:%02d:0%2.4f", int(hours), int(minutes), seconds)
	}
	for i := 0; i < 3; i++ {
		s = strings.TrimSuffix(s, "0")
	}
	return s
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImpr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, mrand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = mrand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
