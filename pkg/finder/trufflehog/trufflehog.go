package trufflehog

import (
	"encoding/json"
	"fmt"
	"github.com/pantheon-systems/search-secrets/pkg/database/enum/reason"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/pantheon-systems/search-secrets/pkg/finder"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

type (
	TruffleHog struct {
		trufflehogCmd []string
		log           *logrus.Logger
	}
	IncomingFinding struct {
		Branch       string   `json:"branch"`
		Commit       string   `json:"commit"`
		CommitHash   string   `json:"commitHash"`
		Diff         string   `json:"diff"`
		Path         string   `json:"path"`
		PrintDiff    string   `json:"printDiff"`
		Reason       string   `json:"reason"`
		StringsFound []string `json:"stringsFound"`
	}
)

func New(trufflehogCmd []string, log *logrus.Logger) (*TruffleHog, error) {
	return &TruffleHog{
		trufflehogCmd: trufflehogCmd,
		log:           log,
	}, nil
}

func (t *TruffleHog) GetFindings(cloneDir string, useEntropy bool, reasonRegexps *finder.ReasonRegexps, out chan *finder.DriverFinding) error {
	regexpFile, err := ioutil.TempFile("", "regexpFile.*.json")
	if err != nil {
		return err
	}
	//defer os.Remove(regexpFile.Name())

	bytesJSON, err := json.MarshalIndent(reasonRegexps, "", "    ")
	if err != nil {
		return err
	}

	if _, err = regexpFile.Write(bytesJSON); err != nil {
		return err
	}

	// Close the file
	err = regexpFile.Close()

	cmdPieces := append(t.trufflehogCmd, []string{
		"--repo_path", cloneDir,
		"--rules", regexpFile.Name(),
		"--json",
		"--regex",
		"--entropy", strconv.FormatBool(useEntropy),
		"--skip_fs",
		"--branch", "master",
		"--skip_fetch",
		"",
	}...)
	cmd := exec.Command(cmdPieces[0], cmdPieces[1:]...)
	t.log.Debug("Running: " + commandString(cmd))

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	dec := json.NewDecoder(cmdReader)

	if err := cmd.Start(); err != nil {
		return errors.Wrapv(err, "unable to start command")
	}

	for dec.More() {
		var incomingFinding IncomingFinding
		if err := dec.Decode(&incomingFinding); err != nil {
			t.log.Fatal(err)
		}

		r := incomingFinding.Reason
		if r == "High Entropy" {
			r = reason.Entropy{}.New().Value()
		}

		finding := &finder.DriverFinding{
			Branch:       incomingFinding.Branch,
			Commit:       incomingFinding.Commit,
			CommitHash:   incomingFinding.CommitHash,
			Diff:         incomingFinding.Diff,
			Path:         incomingFinding.Path,
			PrintDiff:    incomingFinding.PrintDiff,
			Reason:       r,
			StringsFound: incomingFinding.StringsFound,
		}

		out <- finding
	}

	if err := cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.log.Debugf("truffleHog exited with a %d", exitError.ExitCode())
		} else {
			return errors.Wrapv(err, "unable to wait for command")
		}
	}

	return nil
}

func commandString(cmd *exec.Cmd) string {
	var cmdQuoted []string
	for _, a := range cmd.Args {
		cmdQuoted = append(cmdQuoted, fmt.Sprintf("\"%s\"", a))
	}
	cmdString := strings.Join(cmdQuoted, " ")
	if cmd.Dir != "" {
		cmdString = fmt.Sprintf("(cd %s && %s)", cmd.Dir, cmdString)
	}
	return cmdString
}
