package github

import (
	"context"
	"math"
	"slices"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/Khan/genqlient/graphql"
	"golang.org/x/mod/semver"
)

type GetTagsInput struct {
	Owner    string
	Repo     string
	Limit    int
	PageSize int
}

func GetTags(ctx context.Context, client graphql.Client, in GetTagsInput) ([]Ref, error) {
	var tags []Ref

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
		if len(tags) >= in.Limit {
			break
		}

		resp, err := getTags(ctx, client, in.Owner, in.Repo, 10, after)
		if err != nil {
			return nil, err
		}

		pageInfo := resp.Repository.Refs.PageInfo
		tags = append(tags, resp.Repository.Refs.Nodes...)

		for _, ref := range resp.Repository.Refs.Nodes {
			logger.Info(ref.Name, "target", ref.Target.GetOid())
		}

		if pageInfo.HasNextPage {
			after = pageInfo.EndCursor
		} else {
			break
		}
	}

	slices.SortStableFunc(tags, func(a, b Ref) int {
		return semver.Compare(a.Name, b.Name)
	})

	return tags, nil
}
