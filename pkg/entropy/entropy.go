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

func HasHighEntropy(inputString, charsetName string, entropyThreshold float64) (result bool, err error) {
    var charsetChars string
    charsetChars, err = getCharsetChars(charsetName)
    if err != nil {
        err = errors.WithMessagev(err, "unable to get charset characters for name", charsetName)
        return
    }

    if !isStringOfCharset(inputString, charsetChars) {
        result = false
        return
    }

    entropy := entropy(inputString, charsetChars)
    result = entropy >= entropyThreshold

    return
}

func FindHighEntropyWords(inputString, charsetName string, lengthThreshold int, entropyThreshold float64) (result []*structures.LineRangeValue, err error) {
    var charsetChars string
    charsetChars, err = getCharsetChars(charsetName)
    if err != nil {
        err = errors.WithMessagev(err, "unable to get charset characters for name", charsetName)
        return
    }

    indexRanges := findLongStringsOfCharset(inputString, charsetChars, lengthThreshold)
    for _, inRange := range indexRanges {
        rangeValue := inRange.ExtractValue(inputString)
        entropy := entropy(rangeValue.Value, charsetChars)
        if entropy >= entropyThreshold {
            result = append(result, rangeValue)
        }
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

func isStringOfCharset(input, charsetChars string) (result bool) {
    for _, char := range input {
        charString := string(char)
        if !strings.Contains(charsetChars, charString) {
            return false
        }
    }

    return true
}

func findLongStringsOfCharset(input, charsetChars string, threshold int) (result []*structures.LineRange) {
    var startIndex int
    var currentIndex int

    for i, char := range input {
        currentIndex = i
        charString := string(char)

        if strings.Contains(charsetChars, charString) {
            continue
        }

        if currentIndex-startIndex >= threshold {
            result = append(result, structures.NewLineRange(startIndex, currentIndex))
        }

        startIndex = currentIndex + 1
    }

    if currentIndex-startIndex >= threshold {
        result = append(result, structures.NewLineRange(startIndex, currentIndex+1))
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
