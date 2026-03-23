package git

import (
	"fmt"
	"os/exec"
	"strings"

	"charm.land/log/v2"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func RepoPath() (string, error) {
	path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(path)), nil
}

func RevListAncestryPath(repoPath string, repo *git.Repository, start, end plumbing.ObjectID) ([]plumbing.ObjectID, error) {
	var cmd *exec.Cmd
	{
		bin := "git"
		args := []string{
			"rev-list",
			"--ancestry-path",
			start.String() + ".." + end.String(),
		}
		cmd = exec.Command(bin, args...)
		log.Debug("running command", "bin", bin, "args", strings.Join(args, "\n"))
	}

	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	oids := make([]plumbing.ObjectID, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse hash
		if oid, ok := plumbing.FromHex(line); !ok {
			return nil, fmt.Errorf("failed to convert line %q to objectID", line)
		} else {
			oids = append(oids, oid)
		}

	}

	return oids, nil
}

func ToCommitObj(repo *git.Repository, revs ...plumbing.ObjectID) ([]*object.Commit, error) {
	commits := make([]*object.Commit, 0, len(revs))
	for _, rev := range revs {
		commit, err := repo.CommitObject(rev)
		if err != nil {
			return nil, fmt.Errorf("failed to get commit %s: %w", rev.String(), err)
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func RevParseHEAD(repoPath string) (plumbing.ObjectID, error) {
	var cmd *exec.Cmd
	{
		bin := "git"
		args := []string{
			"rev-parse", "HEAD",
		}
		cmd = exec.Command(bin, args...)
		log.Debug("running command", "bin", bin, "args", strings.Join(args, "\n"))
	}
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return plumbing.ObjectID{}, err
	}

	outputTrimmed := strings.TrimSpace(string(output))
	if oid, ok := plumbing.FromHex(outputTrimmed); !ok {
		return plumbing.ObjectID{}, fmt.Errorf("failed to convert %q to objectID", outputTrimmed)
	} else {
		return oid, nil
	}
}
