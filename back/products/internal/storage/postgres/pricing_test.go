package postgres

import "testing"

func TestComputeClientPrice(t *testing.T) {
	tests := []struct {
		name            string
		basePrice       float64
		discountPercent float64
		want            float64
	}{
		{"no discount", 1000, 0, 1000},
		{"20 percent off", 1000, 20, 800},
		{"33 percent off rounds to two decimals", 100, 33, 67},
		{"full discount", 500, 100, 0},
		{"fractional cents round", 99.99, 10, 89.99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeClientPrice(tt.basePrice, tt.discountPercent)
			if got != tt.want {
				t.Fatalf("computeClientPrice(%v, %v) = %v, want %v", tt.basePrice, tt.discountPercent, got, tt.want)
			}
		})
	}
}
