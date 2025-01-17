package types

import (
	"reflect"
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

func TestSortMovieSlice(t *testing.T) {
	tests := []struct {
		name   string
		movies []*MovieOverviewData
		want   []*MovieOverviewData
	}{
		{
			name: "Test with unsorted movies",
			movies: []*MovieOverviewData{
				{Movie: &Movie{Media: Media{Title: "Z Movie"}}},
				{Movie: &Movie{Media: Media{Title: "A Movie"}}},
			},
			want: []*MovieOverviewData{
				{Movie: &Movie{Media: Media{Title: "A Movie"}}},
				{Movie: &Movie{Media: Media{Title: "Z Movie"}}},
			},
		},
		{
			name: "Test with already sorted movies",
			movies: []*MovieOverviewData{
				{Movie: &Movie{Media: Media{Title: "A Movie"}}},
				{Movie: &Movie{Media: Media{Title: "B Movie"}}},
			},
			want: []*MovieOverviewData{
				{Movie: &Movie{Media: Media{Title: "A Movie"}}},
				{Movie: &Movie{Media: Media{Title: "B Movie"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortMovieSlice(tt.movies)
			if !reflect.DeepEqual(tt.movies, tt.want) {
				t.Errorf("SortMovieSlice() = %v, want %v", tt.movies, tt.want)
			}
		})
	}
}

func TestNewEntry(t *testing.T) {
	type args struct {
		name    string
		watched bool
		comment string
	}
	tests := []struct {
		name string
		args args
		want *Entry
	}{
		{
			name: "Test with valid entry",
			args: args{name: "John Doe", watched: true, comment: "Great movie!"},
			want: &Entry{Name: "John Doe", Watched: true, Comment: []byte("Great movie!")},
		},
		{
			name: "Test with empty comment",
			args: args{name: "Jane Doe", watched: false, comment: ""},
			want: &Entry{Name: "Jane Doe", Watched: false, Comment: []byte("")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEntry(tt.args.name, tt.args.watched, tt.args.comment); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}
