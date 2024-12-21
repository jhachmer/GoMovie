package types

import (
	"github.com/jhachmer/gotocollection/internal/config"
	"testing"
)

func TestMain(m *testing.M) {
	config.Envs.OmdbApiKey = "TESTKEY"
	m.Run()
}

type OmdbMock struct {
}

func (r OmdbMock) SendRequest() (*Movie, error) {
	return &Movie{}, nil
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

func TestOmdbTitleRequest_Validate(t *testing.T) {
	type fields struct {
		title string
		year  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				"The Thing",
				"1982",
			},
			wantErr: false,
		},
		{
			name: "no valid year, too long",
			fields: fields{
				title: "The Thing",
				year:  "17664",
			},
			wantErr: true,
		},
		{
			name: "no valid year, too short",
			fields: fields{
				title: "The Thing",
				year:  "176",
			},
			wantErr: true,
		},
		{
			name: "not valid, empty string",
			fields: fields{
				"",
				"1982",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := OmdbTitleRequest{
				title: tt.fields.title,
				year:  tt.fields.year,
			}
			if err := r.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOmdbIDRequest_Validate(t *testing.T) {
	type fields struct {
		imdbID string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid, 8 digits",
			fields: fields{
				imdbID: "tt12345678",
			},
			wantErr: false,
		},
		{
			name: "valid, 7 digits",
			fields: fields{
				imdbID: "tt1234567",
			},
			wantErr: false,
		},
		{
			name: "not valid, no tt prefix",
			fields: fields{
				imdbID: "12341234",
			},
			wantErr: true,
		},
		{
			name: "not valid, triple t prefix",
			fields: fields{
				imdbID: "ttt12341234",
			},
			wantErr: true,
		},
		{
			name: "not valid, only one t",
			fields: fields{
				imdbID: "t12341234",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := OmdbIDRequest{
				imdbID: tt.fields.imdbID,
			}
			if err := r.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
