package service

import "github.com/ireoluwacodes/subsync/internal/domain"

type WebhookService struct {
	repo domain.WebhookRepository
}

func NewWebhookService(repo domain.WebhookRepository) *WebhookService {
	return &WebhookService{repo: repo}
}
