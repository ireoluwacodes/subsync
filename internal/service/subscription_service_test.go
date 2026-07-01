package service

import "testing"

func TestSubscriptionService_Placeholder(t *testing.T) {
	svc := NewSubscriptionService(nil, nil, nil, nil)
	if svc == nil {
		t.Fatal("expected subscription service")
	}
}
