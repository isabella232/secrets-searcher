package dev

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
)

var (
	Params       *Parameters
	CatchPanic   = true
	RunningTests bool
)

type (
	Parameters struct {
		Filter Filter
	}
	Filter struct {
		Processor string
		Repo      string
		Commit    string
		Path      string
		Line      int
	}
)

// Applies to "entropy" and "pem" processors
func BreakOnDiffLine(line int) {
	if Params.Filter.Line != line {
		return
	}

	print("") // Breakpoint
}

// Applies to "pem" processor
func BreakBeforePemRule(rule string) {
	switch rule {

	case "findMultilineKey":
		print("") // Breakpoint

	case "findSingleLineKey":
		print("") // Breakpoint

	}
}

func BreakBeforeSetterRule(rule string) {
	switch rule {

	case "ShellCmdParamVal":
		print("") // Breakpoint

	case "Generic":
		print("") // Breakpoint

	}
}

// Applies to "regex" processors
func BreakpointInProcessor(path, procName string, line int) {
	if Params.Filter.Processor != "" && Params.Filter.Processor != procName {
		return
	}
	if Params.Filter.Path == "" || Params.Filter.Path != path {
		return
	}
	if Params.Filter.Line != line {
		return
	}

	print("") // Breakpoint
}

// I need a shower after this, what a slow and horrible function.
// Don't use it for anything but dev logs or some such. And debug or trace level only!
// https://github.com/golang/go/issues/18590
func GoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func RegexpTestLink(re *regexp.Regexp, testInput string) (result string) {
	qs := url.Values{}
	qs.Set("regex", re.String())
	qs.Set("testString", testInput)
	u := url.URL{Scheme: "https", Host: "regex101.com", RawQuery: qs.Encode()}
	return u.String()
}

func SprintVal(valInput interface{}, msg string) (result string) {
	return
	val, ok := valInput.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(valInput)
	}

	addr := "[noaddr]"
	if val.CanAddr() {
		addr = fmt.Sprintf("UA==%d", val.UnsafeAddr())
	} else {
		addr = fmt.Sprintf("UA=&%d", val.Elem().UnsafeAddr())
	}

	poin := "[noptr]"
	if val.Type().Kind() == reflect.Ptr {
		poin = fmt.Sprintf("PR==%d", val.Pointer())
	} else {
		poin = fmt.Sprintf("PR=*%d", val.Addr().Pointer())
	}

	result += fmt.Sprintf("%-16s %-16s %-30s %-10s %s\n", addr, poin, val.Type().String(), val.Kind().String(), msg)

	return
}
func PrintVal(val interface{}, msg string) {
	fmt.Print(SprintVal(val, msg))
}
