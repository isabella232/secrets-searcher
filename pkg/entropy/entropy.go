package entropy

import (
	"math"
	"strings"

	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
)

const (
	Base64CharsetName = "base64"
	HexCharsetName    = "hex"

	base64Charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	hexCharset    = "1234567890abcdefABCDEF"
)

type Result struct {
	LineRange *manip.LineRangeValue
	Entropy   float64
}

func Entropy(input, charsetChars string) (result float64) {
	if input == "" {
		return 0
	}
	inputLen := len(input)
	for _, charsetChar := range charsetChars {
		px := float64(strings.Count(input, string(charsetChar))) / float64(inputLen)
		if px > 0 {
			result += -px * math.Log2(px)
		}
	}

	return
}

func AgainstCharset(inputString, charsetName string) (result float64) {
	return Entropy(inputString, getCharsetChars(charsetName))
}

func FindHighEntropyWords(inputString, charsetName string, lengthThreshold int, entropyThreshold float64) (result []*Result) {
	charsetChars := getCharsetChars(charsetName)

	indexRanges := findLongStringsOfCharset(inputString, charsetChars, lengthThreshold)
	for _, inRange := range indexRanges {
		rangeValue := inRange.ExtractValue(inputString)
		entropy := Entropy(rangeValue.Value, charsetChars)
		if entropy >= entropyThreshold {
			result = append(result, &Result{
				LineRange: rangeValue,
				Entropy:   entropy,
			})
		}
	}

	return
}

func ValidCharsets() []string {
	return []string{Base64CharsetName, HexCharsetName}
}

func findLongStringsOfCharset(input, charsetChars string, threshold int) (result []*manip.LineRange) {
	var startIndex int
	var currentIndex int

	for i, char := range input {
		currentIndex = i
		charString := string(char)

		if strings.Contains(charsetChars, charString) {
			continue
		}

		if currentIndex-startIndex >= threshold {
			result = append(result, manip.NewLineRange(startIndex, currentIndex))
		}

		startIndex = currentIndex + 1
	}

	if currentIndex-startIndex >= threshold {
		result = append(result, manip.NewLineRange(startIndex, currentIndex+1))
	}

	return
}

func getCharsetChars(charsetName string) (result string) {
	switch charsetName {
	case Base64CharsetName:
		result = base64Charset
	case HexCharsetName:
		result = hexCharset
	default:
		panic("unknown charset name: " + charsetName)
	}
	return
}
