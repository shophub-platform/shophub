package models

import "testing"

func TestShopReplicas(t *testing.T) {
	tests := []struct {
		name string
		av   Availability
		want int
	}{
		{"standard -> 2", AvailabilityStandard, 2},
		{"high -> 3", AvailabilityHigh, 3},
		{"prazno -> 2", Availability(""), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (Shop{Availability: tt.av}).Replicas(); got != tt.want {
				t.Errorf("Replicas() = %d, očekivano %d", got, tt.want)
			}
		})
	}
}
