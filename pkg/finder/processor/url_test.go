package processor_test

import (
    "fmt"
    . "github.com/pantheon-systems/search-secrets/pkg/finder/processor"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/require"
    "regexp"
    "testing"
)

const (
    encoded0 = "encoded0TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4="
    encoded1 = "encoded1QsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZX"
)

var (
    encoded0LineWhitelist        = structures.NewRegexpSetFromStringsMustCompile([]string{`encoded0`})
    encoded0LineBackrefWhitelist = structures.NewRegexpSetFromStringsMustCompile([]string{
        `me:(` + regexp.QuoteMeta(encoded0) + `)`,
    })
    encoded0LineWrongBackrefWhitelist = structures.NewRegexpSetFromStringsMustCompile([]string{
        `me(:` + regexp.QuoteMeta(encoded0) + `)`,
    })
    encoded0LineWholeBackrefWhitelist = structures.NewRegexpSetFromStringsMustCompile([]string{
        `(` + regexp.QuoteMeta(encoded0) + `)`,
    })
    encoded0URL  = fmt.Sprintf("https://me:%s@example.com/path", encoded0)
    encoded0Line = fmt.Sprintf("This is at the beginning of the string %s This is at the end of the string", encoded0URL)
    log          = logrus.NewEntry(logrus.New())
)

func TestEntropyInURLPath(t *testing.T) {
    url := fmt.Sprintf("https://example.com/path/anotherpath/%s/endpath", encoded0)
    input := fmt.Sprintf("This is at the beginning of the string %s This is at the end of the string", url)
    subject := NewURLProcessor(nil)

    // Fire
    response, _, err := subject.FindInLine(input, log)

    require.NoError(t, err)
    require.NotNil(t, response)
    require.Len(t, response, 1)
    require.Equal(t, encoded0, response[0].Secret.Value)
    require.Equal(t, encoded0, response[0].LineRange.ExtractValue(input).Value)
}

func TestMultipleEntropyValuesInURLPath(t *testing.T) {
    url := fmt.Sprintf("https://example.com/path/anotherpath/%s/inbetween/%s/endpath", encoded0, encoded1)
    input := fmt.Sprintf("This is at the beginning of the string %s This is at the end of the string", url)
    subject := NewURLProcessor(nil)

    // Fire
    response, _, err := subject.FindInLine(input, log)

    require.NoError(t, err)
    require.NotNil(t, response)
    require.Len(t, response, 2)
    require.Equal(t, encoded0, response[0].Secret.Value)
    require.Equal(t, encoded0, response[0].LineRange.ExtractValue(input).Value)
    require.Equal(t, encoded1, response[1].Secret.Value)
    require.Equal(t, encoded1, response[1].LineRange.ExtractValue(input).Value)
}

func TestEntropyAsURLPath(t *testing.T) {
    url := fmt.Sprintf("https://example.com/%s", encoded0)
    input := fmt.Sprintf("This is at the beginning of the string %s This is at the end of the string", url)
    subject := NewURLProcessor(nil)

    // Fire
    response, _, err := subject.FindInLine(input, log)

    require.NoError(t, err)
    require.NotNil(t, response)
    require.Len(t, response, 1)
    require.Equal(t, encoded0, response[0].Secret.Value)
    require.Equal(t, encoded0, response[0].LineRange.ExtractValue(input).Value)
}

func TestEntropyAtStartOfURLPath(t *testing.T) {
    url := fmt.Sprintf("https://example.com/%s/path/anotherpath/", encoded0)
    input := fmt.Sprintf("This is at the beginning of the string %s This is at the end of the string", url)
    subject := NewURLProcessor(nil)

    // Fire
    response, _, err := subject.FindInLine(input, log)

    require.NoError(t, err)
    require.NotNil(t, response)
    require.Len(t, response, 1)
    require.Equal(t, encoded0, response[0].Secret.Value)
    require.Equal(t, encoded0, response[0].LineRange.ExtractValue(input).Value)
}

func TestPasswordInURLPath(t *testing.T) {
    subject := NewURLProcessor(nil)

    // Fire
    response, _, err := subject.FindInLine(encoded0Line, log)

    require.NoError(t, err)
    require.NotNil(t, response)
    require.Len(t, response, 1)
    require.Equal(t, encoded0, response[0].Secret.Value)
    require.Equal(t, encoded0, response[0].LineRange.ExtractValue(encoded0Line).Value)
}

func TestWhitelistedLineWithBackrefPasswordInURLPath(t *testing.T) {
    subject := NewURLProcessor(&encoded0LineBackrefWhitelist)

    // Fire
    response, _, err := subject.FindInLine(encoded0Line, log)

    require.NoError(t, err)
    require.Nil(t, response)
}

func TestWhitelistedLineWithWrongBackrefPasswordInURLPath(t *testing.T) {
    subject := NewURLProcessor(&encoded0LineWrongBackrefWhitelist)

    // Fire
    response, _, err := subject.FindInLine(encoded0Line, log)

    require.NoError(t, err)
    require.Len(t, response, 1)
}

func TestWhitelistedLineWithWholeBackrefPasswordInURLPath(t *testing.T) {
    subject := NewURLProcessor(&encoded0LineWholeBackrefWhitelist)

    // Fire
    response, _, err := subject.FindInLine(encoded0Line, log)

    require.NoError(t, err)
    require.Nil(t, response)
}

func TestWhitelistedLinePasswordInURLPath(t *testing.T) {
    subject := NewURLProcessor(&encoded0LineWhitelist)

    // Fire
    response, _, err := subject.FindInLine(encoded0Line, log)

    require.NoError(t, err)
    require.Nil(t, response)
}

func TestTemplateVarPasswordInURLPath(t *testing.T) {
    pass := "{PASSWORD}"
    url := fmt.Sprintf("https://me:%s@example.com/path", pass)
    input := fmt.Sprintf("This is at the beginning of the string %s This is at the end of the string", url)
    subject := NewURLProcessor(nil)

    // Fire
    response, _, err := subject.FindInLine(input, log)

    require.NoError(t, err)
    require.Nil(t, response)
}
