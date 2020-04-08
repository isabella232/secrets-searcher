package entropy

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "math"
    "strings"
)

const (
    Base64CharsetName = "base64"
    HexCharsetName    = "hex"

    base64Charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
    hexCharset    = "1234567890abcdefABCDEF"
)

func FindHighEntropyWords(unputString, charsetName string, lengthThreshold int, entropyThreshold float64) (result []structures.LineRange) {
    charsetChars, err := getCharsetChars(charsetName)
    if err != nil {
        return
    }

    indexRanges := findLongStringsOfCharset(unputString, charsetChars, lengthThreshold)
    for _, inRange := range indexRanges {
        word := inRange.GetStringFrom(unputString)
        entropy := entropy(word, charsetChars)
        if entropy >= entropyThreshold {
            result = append(result, inRange)
        }
    }

    return
}

func findLongStringsOfCharset(input, charsetChars string, threshold int) (result []structures.LineRange) {
    var startIndex int
    var currentIndex int

    for i, char := range input {
        currentIndex = i
        charString := string(char)

        if strings.Contains(charsetChars, charString) {
            continue
        }

        if currentIndex-startIndex >= threshold {
            result = append(result, structures.LineRange{StartIndex: startIndex, EndIndex: currentIndex})
        }

        startIndex = currentIndex + 1
    }

    if currentIndex-startIndex >= threshold {
        result = append(result, structures.LineRange{StartIndex: startIndex, EndIndex: currentIndex + 1})
    }

    return
}

func entropy(input, charsetChars string) (result float64) {
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

func getCharsetChars(charsetName string) (result string, err error) {
    switch charsetName {
    case Base64CharsetName:
        result = base64Charset
    case HexCharsetName:
        result = hexCharset
    default:
        err = errors.Errorv("unknown charset name", charsetName)
    }
    return
}
