package service

import "testing"

func TestDunningService_Placeholder(t *testing.T) {
	svc := NewDunningService()
	if svc == nil {
		t.Fatal("expected dunning service")
	}
}
