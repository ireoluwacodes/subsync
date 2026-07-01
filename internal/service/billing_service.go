package service

import "github.com/ireoluwacodes/subsync/internal/domain"

type BillingService struct {
	invoices domain.InvoiceRepository
}

func NewBillingService(invoices domain.InvoiceRepository) *BillingService {
	return &BillingService{invoices: invoices}
}
