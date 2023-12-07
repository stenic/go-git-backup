package app

import (
	"context"

	"github.com/stenic/go-git-backup/pkg/model"
)

type Platform struct {
	Name         string
	Organisation string
}

type Collector interface {
	GetRepositories(ctx context.Context, org string) ([]model.Repository, error)
}
