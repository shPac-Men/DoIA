package service

import (
	"AwsProj/internal/app/repository"
)

// internal/app/service/mixing_service.go
type MixingService struct {
	repo repository.Repository
}

func NewMixingService(repo *repository.Repository) *MixingService {
	return &MixingService{repo: *repo}
}
