package manip

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

func CountRunes(input string, r rune) (result int) {
	for _, c := range input {
		if c == r {
			result++
		}
	}
	return
}

func Truncate(s string, i int) string {
	if len(s) < i {
		return s
	}
	if utf8.ValidString(s[:i]) {
		return s[:i]
	}
	// The omission.
	// In reality, a rune can have 1-4 bytes width (not 1 or 2)
	return s[:i+1] // or i-1
}

var lineBreakRe = regexp.MustCompile(`\r?\n`)

func MakeOneLine(s, repl string) (result string) {
	return lineBreakRe.ReplaceAllString(s, repl)
}
