package root

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"charm.land/log/v2"
	"github.com/Khan/genqlient/graphql"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffval"

	"github.com/cakehappens/go-release-please/internal/github"
)

type Config struct {
	Stdout      io.Writer
	Stderr      io.Writer
	GithubToken string
	Verbose     bool
	Flags       *ff.FlagSet
	Command     *ff.Command
}

func New(stdout, stderr io.Writer) *Config {
	var cfg Config
	cfg.Stdout = stdout
	cfg.Stderr = stderr
	cfg.Flags = ff.NewFlagSet("root")
	cfg.Flags.StringVar(&cfg.GithubToken, 0, "github-token", "", "GITHUB_TOKEN")
	cfg.Flags.AddFlag(ff.FlagConfig{
		ShortName: 'v',
		LongName:  "verbose",
		Value:     ffval.NewValue(&cfg.Verbose),
		Usage:     "log verbose output",
		NoDefault: true,
	})
	cfg.Command = &ff.Command{
		Name:      "go-release-please",
		ShortHelp: "creates github releases",
		Usage:     "go-release-please [FLAGS] <SUBCOMMAND> ...",
		Flags:     cfg.Flags,
		Exec:      cfg.Exec,
	}
	return &cfg
}

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

const (
	LabelAutoReleasePending   = "autorelease: pending"
	LabelAutoReleasePublished = "autorelease: published"
)

func (cfg *Config) Exec(ctx context.Context, args []string) error {
	httpClient := http.Client{
		Transport: &authedTransport{
			key:     cfg.GithubToken,
			wrapped: http.DefaultTransport,
		},
	}
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", &httpClient)
	log.WithPrefix("📝 ").Info(
		"Getting Pull Requests with label...",
		"label", LabelAutoReleasePending)

	owner := "googleapis"
	repo := "release-please"
	_, err := github.GetPullRequests(ctx, graphqlClient, github.GetPullRequestsInput{
		Owner:  owner,
		Repo:   repo,
		Labels: []string{LabelAutoReleasePending},
	})
	if err != nil {
		return fmt.Errorf("getting pull requests: %w", err)
	}

	// get Releases

	log.WithPrefix("🚀 ").Info("Getting Releases...")
	_, err = github.GetReleases(ctx, graphqlClient, github.GetReleasesInput{
		Owner: owner,
		Repo:  repo,
	})

	// get Tags
	log.WithPrefix("🏷️ ").Info("Getting Tags...")
	_, err = github.GetTags(ctx, graphqlClient, github.GetTagsInput{
		Owner: owner,
		Repo:  repo,
	})

	return nil
}
