package cmd

import (
	"github.com/pantheon-systems/search-secrets/pkg/app"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	searchCmd = &cobra.Command{
		Use:   "search",
		Short: "Search repositories for all resources across all configured contexts",
		Run:   runSearch,
	}
)

func init() {
	initSearchArgs()
	rootCmd.AddCommand(searchCmd)
}

func initSearchArgs() {
	flags := searchCmd.LocalFlags()

	flags.String(
		"github-token",
		"",
		"GitHub API token.")

	flags.String(
		"output-dir",
		"./output",
		"Output directory.")

	flags.String(
		"organization",
		"",
		"Organization to search.")

	flags.StringSlice(
		"trufflehog-cmd",
		[]string{"./thog.sh"},
		"TruffleHog command.")

	flags.StringSlice(
		"repos",
		[]string{},
		"Only search these repos.")

	flags.StringSlice(
		"reasons",
		[]string{},
		"Only show these reasons.")

	flags.Bool(
		"skip-entropy",
		false,
		"Use every reason except for \"entropy\". If \"reasons\" is passed, this argument is ignored.")
}

func runSearch(*cobra.Command, []string) {
	githubToken := vpr.GetString("github-token")
	outputDir := vpr.GetString("output-dir")
	organization := vpr.GetString("organization")
	truffleHogCmd := vpr.GetStringSlice("trufflehog-cmd")
	reposFilter := vpr.GetStringSlice("repos")
	reasonsFilter := vpr.GetStringSlice("reasons")
	skipEntropy := vpr.GetBool("skip-entropy")

	// Validate
	if organization == "" {
		errors.Fatal(log, errors.New("organization is required"))
	}
	if githubToken == "" {
		errors.Fatal(log, errors.New("github-token is required"))
	}

	search, err := app.NewSearch(githubToken, organization, outputDir, truffleHogCmd, reposFilter, reasonsFilter, skipEntropy, log)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "unable to create search app"))
	}

	if err := search.Execute(); err != nil {
		log.Fatal(errors.WithMessage(err, "unable to execute search app"))
	}
}
