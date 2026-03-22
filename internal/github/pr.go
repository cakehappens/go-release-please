package github

import (
	"context"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/Khan/genqlient/graphql"
	"github.com/samber/lo"

	"github.com/cakehappens/go-release-please/internal/tx"
)

type pullRequestTx struct {
	doFunc   func(ctx context.Context) error
	undoFunc func(ctx context.Context) error

	prId      *string
	url       *string
	permaLink *string
	number    *int
}

func (t *pullRequestTx) Do(ctx context.Context) error {
	if t.doFunc == nil {
		return nil
	}

	return t.doFunc(ctx)
}

func (t *pullRequestTx) Undo(ctx context.Context) error {
	if t.undoFunc == nil {
		return nil
	}

	return t.undoFunc(ctx)
}

type PullRequestInput struct {
	Repo        string
	BaseRefName string
	HeadRefName string
	Title       string
	Body        string
}

func NewPullRequestTx(client graphql.Client, in PullRequestInput) tx.Transacter {
	t := &pullRequestTx{}

	t.doFunc = func(ctx context.Context) error {
		resp, err := CreatePullRequest(ctx, client, in.Repo, in.BaseRefName, in.HeadRefName, in.Title, in.Body)
		if err != nil {
			return fmt.Errorf("creating pull request: %w", err)
		}

		pr := resp.GetCreatePullRequest().PullRequest
		t.prId = &pr.Id
		t.number = &pr.Number
		t.permaLink = &pr.Permalink
		t.url = &pr.Url

		return nil
	}

	t.undoFunc = func(ctx context.Context) error {
		if t.prId != nil {
			_, err := ClosePullRequest(ctx, client, *t.prId)
			if err != nil {
				return fmt.Errorf("closing pull request: %w", err)
			}
		}

		return nil
	}

	return t
}

type GetPullRequestsInput struct {
	Owner       string
	Repo        string
	BaseRefName string
	Labels      []string
}

func GetPullRequests(ctx context.Context, client graphql.Client, in GetPullRequestsInput) ([]PullRequest, error) {
	var prs []PullRequest

	logger := log.Default().With()
	styles := log.DefaultStyles()
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("  |").
		Foreground(lipgloss.Lighten(lipgloss.Black, 0.5))
	logger.SetStyles(styles)

	baseRefName := in.BaseRefName
	if baseRefName == "" {
		baseRefName = "main"
	}

	prAfter := ""
	for {
		resp, err := getPullRequests(ctx, client, in.Owner, in.Repo, baseRefName, in.Labels, prAfter)
		if err != nil {
			return nil, err
		}

		pageInfo := resp.Repository.PullRequests.PageInfo
		prs = append(prs, resp.Repository.PullRequests.Nodes...)

		for _, pr := range resp.Repository.PullRequests.Nodes {
			labels := lo.Map(pr.Labels.Nodes, func(item Label, index int) string {
				return item.Name
			})
			logger.Info(
				fmt.Sprintf("#%d %s", pr.Number, pr.Title),
				"labels", strings.Join(labels, "\n"),
			)
		}

		if pageInfo.HasNextPage {
			prAfter = pageInfo.EndCursor
		} else {
			break
		}
	}

	return prs, nil
}
