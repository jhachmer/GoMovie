package api

import (
	"testing"
)

func TestRating_String(t *testing.T) {
	tests := []struct {
		name string
		r    Rating
		want string
	}{
		{
			name: "Test with valid rating",
			r:    Rating{Source: "Internet", Value: "8.5"},
			want: "8.5",
		},
		{
			name: "Test with empty rating",
			r:    Rating{Source: "Internet", Value: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("Rating.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
