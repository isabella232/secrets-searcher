package search_test

import (
	"testing"

	"github.com/pantheon-systems/search-secrets/pkg/builtin"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/search"
	"github.com/pantheon-systems/search-secrets/pkg/search/searchtest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type targetTest struct {
	coreTarget     builtin.TargetName
	key            string
	val            string
	expMatchResult search.TargetMatchResult
}

var log = logg.NewLogrusLogg(logrus.New())

func runTargetTest(t *testing.T, tt targetTest) {
	subject := searchtest.CoreTarget(tt.coreTarget)

	// Fire
	_, matchResult := subject.Matches(tt.key, tt.val, log)

	assert.Equal(t, tt.expMatchResult.String(), matchResult.String())
}

func TestTarget_ExcludePasswordPolicy(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.Passwords,
		key:            "passwordPolicy",
		val:            "shhhshhhshhhshhhshhhshhhshhhshhhshhhshhh",
		expMatchResult: search.KeyExcluded,
	})
}

func TestTarget_ExcludeFormTokens(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.APIKeysAndTokens,
		key:            "form_token",
		val:            "shhhshhhshhhshhhshhhshhhshhhshhhshhhshhh",
		expMatchResult: search.KeyExcluded,
	})
}

func TestTarget_ExcludeKeyss(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.Passwords,
		key:            "/renewing-america/2012/3/15/morning-brief-senate-passes-highway-bill-focus-turns-to-house",
		val:            "shhhshhhshhhshhhshhhshhhshhhshhhshhhshhh",
		expMatchResult: search.KeyExcluded,
	})
}

func TestTarget_ExcludeKeysEndingWithFile(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.Passwords,
		key:            "passwordPath",
		val:            "shhhshhhshhhshhhshhhshhhshhhshhhshhhshhh",
		expMatchResult: search.KeyExcluded,
	})
}

func TestTarget_ExcludeFilePaths(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.Passwords,
		key:            "password",
		val:            "the/path/to/password.json",
		expMatchResult: search.ValFilePath,
	})
}

func TestTarget_ExcludeFilePaths2(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.Passwords,
		key:            "password",
		val:            "./the/path/path/path",
		expMatchResult: search.ValFilePath,
	})
}

func TestTarget_ExcludeFilePathsNoDirsShouldHit(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.Passwords,
		key:            "password",
		val:            "the-path-to-password.json",
		expMatchResult: search.ValFilePath,
	})
}

func TestTarget_DontExcludeAsPath(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.APIKeysAndTokens,
		key:            "token",
		val:            "pa-07381fe0-4587-4ca4-9d9c-73a00bcbe869",
		expMatchResult: search.Match,
	})
}

func TestTarget_NoEntropy(t *testing.T) {
	runTargetTest(t, targetTest{
		coreTarget:     builtin.APIKeysAndTokens,
		key:            "token",
		val:            "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		expMatchResult: search.ValEntropy,
	})
}
