package root

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"charm.land/log/v2"
	"github.com/Khan/genqlient/graphql"
	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffval"

	"github.com/cakehappens/go-release-please/internal/git"
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
	LabelKey = "label"

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
		LabelKey, LabelAutoReleasePending)

	owner := "googleapis"
	repoShort := "release-please"
	prs, err := github.GetPullRequests(ctx, graphqlClient, github.GetPullRequestsInput{
		Owner:  owner,
		Repo:   repoShort,
		Labels: []string{LabelAutoReleasePending},
		Limit:  20,
	})
	if err != nil {
		return fmt.Errorf("getting pull requests: %w", err)
	}

	if len(prs) == 0 {
		log.Info("No pull requests found with", LabelKey, LabelAutoReleasePending)
		log.Info("Desire: Create a new pull request for the next release (if applicable)")
	}

	// get Releases

	log.WithPrefix("🚀 ").Info("Getting Releases...")
	_, err = github.GetReleases(ctx, graphqlClient, github.GetReleasesInput{
		Owner: owner,
		Repo:  repoShort,
		Limit: 20,
	})

	// get Tags
	log.WithPrefix("🏷️ ").Info("Getting Tags...")
	tags, err := github.GetTags(ctx, graphqlClient, github.GetTagsInput{
		Owner: owner,
		Repo:  repoShort,
		Limit: 20,
	})

	// get commits
	latest := tags[len(tags)-1]
	targetOid := latest.Target.GetOid()
	log.Info("found latest tag", "tag", latest.Name, "target", targetOid)
	latestOid, ok := plumbing.FromHex(targetOid)
	if !ok {
		return fmt.Errorf("failed to convert tagged commit %q to objectID", targetOid)
	}

	repoPath, err := git.RepoPath()
	if err != nil {
		return fmt.Errorf("getting root of repo")
	}

	currentHead, err := git.RevParseHEAD(repoPath)
	if err != nil {
		return fmt.Errorf("getting HEAD")
	}

	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("opening local repo at %q: %w", repoPath, err)
	}

	history, err := git.RevListAncestryPath(repoPath, repo, latestOid, currentHead)
	if err != nil {
		return fmt.Errorf("performing git rev-list --ancestry-path: %w", err)
	}

	commitHistory, err := git.ToCommitObj(repo, history...)
	if len(commitHistory) == 0 {
		log.Info("No ancestral commits between current HEAD and latest tag",
			"HEAD", currentHead,
			"tag-name", latest.Name,
			"tag-target", latestOid,
		)
	}

	for _, c := range commitHistory {
		convCommit := git.ParseConventionalCommit(c.Message)
		commitHashShort := c.Hash.String()[0:8]

		commitLogger := log.With("valid", convCommit.IsValid())
		if convCommit.IsValid() {
			commitLogger.Info(commitHashShort,
				"type", convCommit.Type,
				"scope", convCommit.Scope,
				"breaking", convCommit.Breaking,
				"description", convCommit.Description,
			)
		} else {
			commitLogger.Info(commitHashShort,
				"header", convCommit.RAWHeader,
			)
		}
	}

	return nil
}
