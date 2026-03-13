package api

import (
	"reflect"
	"testing"
)

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

func TestSortMovieSlice(t *testing.T) {
	tests := []struct {
		name   string
		movies []*MovieInfoData
		want   []*MovieInfoData
	}{
		{
			name: "Test with unsorted movies",
			movies: []*MovieInfoData{
				{Movie: &Movie{Title: "Z Movie"}},
				{Movie: &Movie{Title: "A Movie"}},
			},
			want: []*MovieInfoData{
				{Movie: &Movie{Title: "A Movie"}},
				{Movie: &Movie{Title: "Z Movie"}},
			},
		},
		{
			name: "Test with already sorted movies",
			movies: []*MovieInfoData{
				{Movie: &Movie{Title: "A Movie"}},
				{Movie: &Movie{Title: "B Movie"}},
			},
			want: []*MovieInfoData{
				{Movie: &Movie{Title: "A Movie"}},
				{Movie: &Movie{Title: "B Movie"}},
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
