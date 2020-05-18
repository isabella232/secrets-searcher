package regex_test

import (
	"testing"

	"github.com/pantheon-systems/search-secrets/pkg/app/build"
	"github.com/pantheon-systems/search-secrets/pkg/builtin"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/search/searchtest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type processorTest struct {
	coreProcessor builtin.ProcessorName
	line          string
	expMatch      bool
	expSecret     string
	expContext    string
}

var log = logg.NewLogrusLogg(logrus.New())

func runProcessorTest(t *testing.T, tt processorTest) {
	procConfig := builtin.ProcessorConfig(tt.coreProcessor)
	subject := build.ProcRegex(procConfig.Name, &procConfig.RegexProcessorConfig, log)
	job := &searchtest.LineProcJobMock{Logger: log}

	// Fire
	err := subject.FindResultsInLine(job, tt.line)

	require.NoError(t, err)
	if !tt.expMatch {
		assert.Len(t, job.LineRanges, 0)
		return
	}

	assert.Len(t, job.LineRanges, 1)
	assert.Equal(t, tt.expSecret, job.LineRanges[0].ExtractValue(tt.line).Value)
	assert.Equal(t, tt.expContext, job.ContextLineRanges[0].ExtractValue(tt.line).Value)
}

func TestProcessor_URLPasswordRegex_Long(t *testing.T) {
	runProcessorTest(t, processorTest{
		coreProcessor: builtin.URLPasswordRegex,
		line:          `"https://app.surveygizmo.com/bu", "https://app.surveygizmo.com/bu", WebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3","muhammad17@example.net`,
		expMatch:      false,
	})
}

func TestProcessor_URLPasswordRegex_PHPVariablePassword(t *testing.T) {
	runProcessorTest(t, processorTest{
		coreProcessor: builtin.URLPasswordRegex,
		line:          `https://$GITHUB_USER:$GITHUB_TOKEN@github.com`,
		expMatch:      false,
	})
}
