package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stenic/go-git-backup/pkg/github"
	"github.com/stenic/go-git-backup/pkg/model"

	"github.com/NoUseFreak/go-parallel"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func init() {
	wd, _ := os.Getwd()
	viper.SetDefault("repos.dir", filepath.Join(wd, "repos"))
	viper.SetDefault("bundles.dir", filepath.Join(wd, "bundles"))
	viper.SetDefault("github.organisation", "")
	viper.SetDefault("tune.parallel", 5)
}

func Run(ctx context.Context, platform Platform) error {
	start := time.Now()

	token := viper.GetString(platform.Name + ".token")
	repoDir := viper.GetString("repos.dir")
	bundleDir := viper.GetString("bundles.dir")
	logger := logrus.WithField("platform", platform.Name)
	gitAuth := http.BasicAuth{
		Username: "token",
		Password: token,
	}

	var collector Collector
	switch platform.Name {
	case "github":
		collector = github.NewGithubCollector(ctx, token)
	default:
		return errors.New("unknown platform")
	}

	logger.Info("Prepare environment")
	os.RemoveAll(bundleDir)
	os.MkdirAll(bundleDir, 0755)

	logger.Info("Collecting repos in organisation")
	repos, err := collector.GetRepositories(ctx, platform.Organisation)
	if err != nil {
		return err
	}

	logger.Info("Creating bundles")

	input := parallel.Input{}
	for _, repo := range repos {
		input = append(input, repo)
	}
	bundler := Bundler{
		RepoDir:   repoDir,
		BundleDir: bundleDir,
		Logger:    logger,
	}
	p := parallel.Processor{Threads: viper.GetInt("tune.parallel")}
	p.Process(input, func(i interface{}) interface{} {
		repo := i.(model.Repository)
		logger.Infof("Creating bundle: %s", repo.Slug)
		return bundler.Bundle(ctx, repo, gitAuth)
	})

	logger.Info("Gathering stats")
	bundleFiles, _ := filepath.Glob(bundleDir + "/**/*.bundle")
	totalSize := int64(0)
	for _, bundleFile := range bundleFiles {
		info, err := os.Stat(bundleFile)
		if err != nil {
			logger.Error(err)
		}
		totalSize += info.Size()
	}

	logger.WithFields(logrus.Fields{
		"duration":  time.Since(start).Truncate(time.Second).String(),
		"totalSize": humanize.Bytes(uint64(totalSize)),
	}).Warnf("Created bundles for %d/%d repos", len(bundleFiles), len(repos))

	return nil
}
