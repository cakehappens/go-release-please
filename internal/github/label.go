package github

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"

	"github.com/cakehappens/go-release-please/internal/collectionutil"
	"github.com/cakehappens/go-release-please/internal/tx"
)

type pullRequestLabelTx struct {
	doFunc   func(ctx context.Context) error
	undoFunc func(ctx context.Context) error

	prID        *string
	prNumber    int
	labelsAdded []string
}

func (t *pullRequestLabelTx) Do(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (t *pullRequestLabelTx) Undo(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

type PullRequestLabelInput struct {
	Owner    string
	RepoName string
	PrNumber int
	LabelIds []string
}

func NewPullRequestLabelTx(client graphql.Client, in PullRequestLabelInput) tx.Transacter {
	t := &pullRequestLabelTx{}

	t.doFunc = func(ctx context.Context) error {
		resp, err := GetPullRequest(ctx, client, in.Owner, in.RepoName, in.PrNumber)
		if err != nil {
			return fmt.Errorf("getting pull request %d: %w", in.PrNumber, err)
		}

		pr := resp.Repository.PullRequest
		t.prID = &pr.Id
		t.prNumber = pr.Number

		labelsDesired := collectionutil.NewSet(in.LabelIds...)
		labelsToAdd := collectionutil.NewSet[string]()

		for _, label := range pr.Labels.Nodes {
			if !labelsDesired.Contains(label.Id) {
				labelsToAdd.Add(label.Id)
			}
		}

		labelsList := labelsToAdd.List()

		_, err = AddLabels(ctx, client, pr.Id, labelsList)
		if err != nil {
			return fmt.Errorf("adding labels (ids: %s) to PR %d: %w", in.LabelIds, t.prNumber, err)
		}

		t.labelsAdded = labelsList

		return nil
	}

	t.undoFunc = func(ctx context.Context) error {
		if t.prID != nil && len(t.labelsAdded) > 0 {
			_, err := RemoveLabels(ctx, client, *t.prID, t.labelsAdded)
			if err != nil {
				return fmt.Errorf("removing labels (%s) from PR %d: %w", t.labelsAdded, t.prNumber, err)
			}
		}

		return nil
	}

	return t
}

type pullRequestUnlabelerTx struct {
	doFunc   func(ctx context.Context) error
	undoFunc func(ctx context.Context) error

	prID          *string
	prNumber      int
	labelsRemoved []string
}

func (t *pullRequestUnlabelerTx) Do(ctx context.Context) error {
	if t.doFunc == nil {
		return nil
	}

	return t.doFunc(ctx)
}

func (t *pullRequestUnlabelerTx) Undo(ctx context.Context) error {
	if t.undoFunc == nil {
		return nil
	}

	return t.doFunc(ctx)
}

func NewPullRequestUnlabelerTx(client graphql.Client, in PullRequestLabelInput) tx.Transacter {
	t := &pullRequestUnlabelerTx{}

	t.doFunc = func(ctx context.Context) error {
		resp, err := GetPullRequest(ctx, client, in.Owner, in.RepoName, in.PrNumber)
		if err != nil {
			return fmt.Errorf("getting pull request %d: %w", in.PrNumber, err)
		}

		pr := resp.Repository.PullRequest
		t.prID = &pr.Id
		t.prNumber = pr.Number

		labelsDesiredToRemove := collectionutil.NewSet(in.LabelIds...)
		labelsToRemove := collectionutil.NewSet[string]()

		for _, label := range pr.Labels.Nodes {
			if labelsDesiredToRemove.Contains(label.Id) {
				labelsToRemove.Add(label.Id)
			}
		}

		labelsList := labelsToRemove.List()
		_, err = RemoveLabels(ctx, client, *t.prID, labelsList)
		if err != nil {
			return fmt.Errorf("removing labels (%s) to PR %d: %w", labelsList, pr.Number, err)
		}

		t.labelsRemoved = labelsList

		return nil
	}

	t.undoFunc = func(ctx context.Context) error {
		if t.prID != nil && len(t.labelsRemoved) > 0 {
			_, err := AddLabels(ctx, client, *t.prID, t.labelsRemoved)
			if err != nil {
				return fmt.Errorf("adding labels (%s) to PR %d: %w", t.labelsRemoved, t.prNumber, err)
			}
		}

		return nil
	}

	return t
}

type createLabelTx struct {
	doFunc   func(ctx context.Context) error
	undoFunc func(ctx context.Context) error

	labelId *string
}

func (t *createLabelTx) Do(ctx context.Context) error {
	if t.doFunc == nil {
		return nil
	}

	return t.doFunc(ctx)
}

func (t *createLabelTx) Undo(ctx context.Context) error {
	if t.undoFunc == nil {
		return nil
	}

	return t.doFunc(ctx)
}

type CreateLabelInput struct {
	Owner       string
	RepoName    string
	Name        string
	Color       string
	Description string
}

func NewCreateLabelTx(client graphql.Client, in CreateLabelInput) tx.Transacter {
	t := &createLabelTx{}

	t.doFunc = func(ctx context.Context) error {
		repoResp, err := GetRepo(ctx, client, in.Owner, in.RepoName)
		if err != nil {
			return fmt.Errorf("getting repository: %w", err)
		}

		repo := repoResp.Repository

		labelResp, err := CreateLabel(ctx, client, repo.Id, in.Name, in.Color, in.Description)
		if err != nil {
			return fmt.Errorf("creating label: %w", err)
		}

		t.labelId = &labelResp.CreateLabel.Label.Id

		return nil
	}

	t.undoFunc = func(ctx context.Context) error {
		if t.labelId != nil {
			_, err := DeleteLabel(ctx, client, *t.labelId)
			if err != nil {
				return fmt.Errorf("deleting label with ID %s: %w", *t.labelId, err)
			}
		}
		return nil
	}

	return t
}
