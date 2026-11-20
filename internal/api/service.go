package api

import (
	"context"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Pinger(ctx context.Context) error {
	ctxTime, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.repo.Ping(ctxTime)
}
