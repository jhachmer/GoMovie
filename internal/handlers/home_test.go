package handlers

/*
func Test_parseSearchQuery(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want store.SearchParams
	}{
		{
			name: "valid string",
			args: args{query: "genre:horror,thriller;actors:Hans Albers,Keeanu Reeves"},
			want: store.SearchParams{
				Genres: []string{"horror", "thriller"},
				Actors: []string{"Hans Albers", "Keeanu Reeves"},
				Years:  store.YearSearch{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSearchQuery(tt.args.query); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSearchQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
