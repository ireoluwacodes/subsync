package service

import "testing"

func TestBillingService_Placeholder(t *testing.T) {
	svc := NewBillingService(nil)
	if svc == nil {
		t.Fatal("expected billing service")
	}
}
