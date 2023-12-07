package app

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/stenic/go-git-backup/pkg/model"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type BundleReport struct {
	Error error
}
type Bundler struct {
	RepoDir   string
	BundleDir string
	Logger    *logrus.Entry
}

func (b Bundler) Bundle(ctx context.Context, repo model.Repository, gitAuth http.BasicAuth) BundleReport {
	rlog := b.Logger.WithField("repo", repo.Slug)
	target := filepath.Join(b.RepoDir, repo.Slug)

	rlog.Trace("Creating target dir")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return BundleReport{Error: err}
	}
	defer func() {
		rlog.Trace("Removing target dir")
		os.RemoveAll(target)
	}()

	rlog.Trace("Cloning repo")
	gitrepo, err := b.cloneRepo(target, repo.URL, &gitAuth)
	if err != nil {
		switch {
		case errors.Is(err, transport.ErrEmptyRemoteRepository):
			rlog.Warn(err)
		default:
			rlog.Error(err)
		}
		return BundleReport{Error: err}
	}

	if gitrepo != nil {
		rlog.Trace("Fetching repo")
		if err := gitrepo.FetchContext(ctx, &git.FetchOptions{
			Auth: &gitAuth,
		}); err != nil && !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			rlog.Error(err)
			return BundleReport{Error: err}
		}
	}

	rlog.Trace("Creating bundle")
	bundleFile := filepath.Join(b.BundleDir, repo.Slug+".bundle")
	os.Mkdir(filepath.Dir(bundleFile), 0755)
	cmd := exec.CommandContext(ctx, "bash", "-c", "git bundle create "+bundleFile+" --all")
	cmd.Dir = target
	out, err := cmd.CombinedOutput()
	logrus.Trace(string(out))
	if err != nil {
		logrus.Error(err)

	}
	return BundleReport{Error: err}
}

func (b Bundler) cloneRepo(target string, url string, auth *http.BasicAuth) (*git.Repository, error) {
	var gitrepo *git.Repository
	if _, err := os.Stat(target); err == nil {
		if gitrepo, err = git.PlainOpen(target); err != nil {
			return gitrepo, err
		}
	} else {
		if err := os.MkdirAll(target, 0755); err != nil {
			return gitrepo, err
		}
		if gitrepo, err = git.PlainClone(target, false, &git.CloneOptions{
			URL: url,
			// Progress: os.Stdout,
			Auth: auth,
		}); err != nil {
			return gitrepo, err
		}
	}
	return nil, nil
}
