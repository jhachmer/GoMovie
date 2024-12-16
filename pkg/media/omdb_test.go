package media

import (
	"github.com/jhachmer/gotocollection/pkg/config"
	"github.com/jhachmer/gotocollection/pkg/types"
	"testing"
)

func TestMain(m *testing.M) {
	config.Envs.OmdbApiKey = "TESTKEY"
	m.Run()
}

type OmdbMock struct {
}

func (r OmdbMock) SendRequest() (*types.Movie, error) {
	return &types.Movie{}, nil
}

func (r OmdbMock) Validate() error {
	return nil
}

func Test_makeRequestURL(t *testing.T) {
	type args struct {
		r OmdbRequest
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "ID",
			args:    args{r: OmdbIDRequest{imdbID: "tt123123"}},
			want:    "http://www.omdbapi.com/?apikey=TESTKEY&i=tt123123",
			wantErr: false,
		},
		{
			name: "Title and Year",
			args: args{r: OmdbTitleRequest{
				title: "TestMovie",
				year:  "1984",
			}},
			want:    "http://www.omdbapi.com/?apikey=TESTKEY&t=TestMovie&y=1984",
			wantErr: false,
		},
		{
			name:    "Default case",
			args:    args{r: OmdbMock{}},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeRequestURL(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeRequestURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("makeRequestURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
