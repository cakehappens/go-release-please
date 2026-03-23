package github

import (
	"context"
	"fmt"
	"math"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/Khan/genqlient/graphql"

	"github.com/cakehappens/go-release-please/internal/tx"

	"github.com/google/go-github/v84/github"
)

type releaseTx struct {
	doFunc   func(ctx context.Context) error
	undoFunc func(ctx context.Context) error

	releaseID *int64
}

func (t *releaseTx) Do(ctx context.Context) error {
	if t.doFunc == nil {
		return nil
	}

	return t.doFunc(ctx)
}

func (t *releaseTx) Undo(ctx context.Context) error {
	if t.undoFunc == nil {
		return nil
	}

	return t.undoFunc(ctx)
}

func NewReleaseTx(client *github.Client, owner, repo string, release *github.RepositoryRelease) tx.Transacter {
	t := &releaseTx{}

	t.doFunc = func(ctx context.Context) error {
		rel, _, err := client.Repositories.CreateRelease(ctx, owner, repo, release)
		if err != nil {
			return fmt.Errorf("creating release: %w", err)
		}
		t.releaseID = rel.ID
		return nil
	}

	t.undoFunc = func(ctx context.Context) error {
		if t.releaseID != nil {
			_, err := client.Repositories.DeleteRelease(ctx, owner, repo, *t.releaseID)
			if err != nil {
				return fmt.Errorf("deleting release: %w", err)
			}
		}

		return nil
	}

	return t
}

type GetReleasesInput struct {
	Owner    string
	Repo     string
	Limit    int
	PageSize int
}

func GetReleases(ctx context.Context, client graphql.Client, in GetReleasesInput) ([]Release, error) {
	var releases []Release

	if in.Limit == 0 {
		in.Limit = math.MaxInt
	}

	if in.PageSize == 0 {
		in.PageSize = 100
	}

	logger := log.Default().With()
	styles := log.DefaultStyles()
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("  |").
		Foreground(lipgloss.Lighten(lipgloss.Black, 0.5))
	logger.SetStyles(styles)

	after := ""
	for {
		if len(releases) >= in.Limit {
			break
		}

		resp, err := getReleases(ctx, client, in.Owner, in.Repo, in.PageSize, after)
		if err != nil {
			return nil, err
		}

		pageInfo := resp.Repository.Releases.PageInfo
		releases = append(releases, resp.Repository.Releases.Nodes...)

		for _, rel := range resp.Repository.Releases.Nodes {
			logger.Info(rel.Name, "tag", rel.TagName)
		}

		if pageInfo.HasNextPage {
			after = pageInfo.EndCursor
		} else {
			break
		}
	}

	return releases, nil
}
