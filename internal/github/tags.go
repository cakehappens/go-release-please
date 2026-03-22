package github

import (
	"context"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/Khan/genqlient/graphql"
)

type GetTagsInput struct {
	Owner string
	Repo  string
}

func GetTags(ctx context.Context, client graphql.Client, in GetTagsInput) ([]Ref, error) {
	var tags []Ref

	logger := log.Default().With()
	styles := log.DefaultStyles()
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("  |").
		Foreground(lipgloss.Lighten(lipgloss.Black, 0.5))
	logger.SetStyles(styles)

	after := ""
	for {
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

	return tags, nil
}
