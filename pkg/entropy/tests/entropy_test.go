package entropy_test

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/entropy"
    "github.com/stretchr/testify/require"
    "testing"
)

func TestEntropyInString(t *testing.T) {
    encoded := "TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4="
    input := fmt.Sprintf("This is at the beginning of the string %s This is at the end of the string", encoded)

    // Fire
    response := entropy.FindEntropyBase64(input)

    require.NotNil(t, response)
    require.Len(t, response, 1)
    require.Equal(t, encoded, response[0].GetStringFrom(input))
}

func TestEntropyEndOfString(t *testing.T) {
    encoded := "TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4="
    input := fmt.Sprintf("This is at the beginning of the string %s", encoded)

    // Fire
    response := entropy.FindEntropyBase64(input)

    require.NotNil(t, response)
    require.Len(t, response, 1)
    require.Equal(t, encoded, response[0].GetStringFrom(input))
}

func TestEntropyBeginningOfString(t *testing.T) {
    encoded := "TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4="
    input := fmt.Sprintf("%s This is at the end of the string", encoded)

    // Fire
    response := entropy.FindEntropyBase64(input)

    require.NotNil(t, response)
    require.Len(t, response, 1)
    require.Equal(t, encoded, response[0].GetStringFrom(input))
}

func TestEntropyWholeString(t *testing.T) {
    encoded := "TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4="

    // Fire
    response := entropy.FindEntropyBase64(encoded)

    require.NotNil(t, response)
    require.Len(t, response, 1)
    require.Equal(t, encoded, response[0].GetStringFrom(encoded))
}
