package domain

import (
	"context"
	"io"
)

type FetchRepo interface {
	Fetch(ctx context.Context, name string) (io.ReadCloser, error)
}

type UseCaseFetch struct {
	repo FetchRepo
}

func NewUseCaseFetch(repo FetchRepo) UseCaseFetch {
	return UseCaseFetch{repo: repo}
}

func (u UseCaseFetch) Run(ctx context.Context, name string) (io.ReadCloser, error) {
	return u.repo.Fetch(ctx, name)
}
