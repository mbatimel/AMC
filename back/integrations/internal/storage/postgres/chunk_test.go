package postgres

import (
	"reflect"
	"testing"
)

func TestChunkRanges(t *testing.T) {
	cases := []struct {
		name string
		n    int
		size int
		want [][2]int
	}{
		{"empty", 0, 500, nil},
		{"single_partial_chunk", 3, 500, [][2]int{{0, 3}}},
		{"exact_multiple", 6, 3, [][2]int{{0, 3}, {3, 6}}},
		{"remainder", 7, 3, [][2]int{{0, 3}, {3, 6}, {6, 7}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := chunkRanges(c.n, c.size)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("chunkRanges(%d,%d) = %v, want %v", c.n, c.size, got, c.want)
			}
		})
	}
}
