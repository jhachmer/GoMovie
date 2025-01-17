package handlers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jhachmer/gotocollection/internal/types"
)

func Test_parseSearchQuery(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name    string
		args    args
		want    types.SearchParams
		wantErr bool
	}{
		{
			name: "valid string",
			args: args{query: "genre:horror,thriller;actors:Hans Albers,Keeanu Reeves"},
			want: types.SearchParams{
				Genres: []string{"horror", "thriller"},
				Actors: []string{"Hans Albers", "Keeanu Reeves"},
				Years:  types.YearSearch{},
			},
			wantErr: false,
		},
		{
			name:    "invalid search type",
			args:    args{query: "invalid:horror"},
			want:    types.SearchParams{},
			wantErr: true,
		},
		{
			name:    "missing colon",
			args:    args{query: "genre horror"},
			want:    types.SearchParams{},
			wantErr: true,
		},
		{
			name: "valid year range",
			args: args{query: "year:1990,2000"},
			want: types.SearchParams{
				Genres: nil,
				Actors: nil,
				Years:  types.YearSearch{StartYear: "1990", EndYear: "2000"},
			},
			wantErr: false,
		},
		{
			name: "valid mixed types",
			args: args{query: "genre:horror;actors:Hans Albers;year:1990,2000"},
			want: types.SearchParams{
				Genres: []string{"horror"},
				Actors: []string{"Hans Albers"},
				Years:  types.YearSearch{StartYear: "1990", EndYear: "2000"},
			},
			wantErr: false,
		},
		{
			name:    "empty query",
			args:    args{query: ""},
			want:    types.SearchParams{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSearchQuery(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSearchQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("parseSearchQuery() = %v, want %v", got, tt.want)
			//}
		})
	}
}
