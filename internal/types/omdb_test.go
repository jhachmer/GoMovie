package types

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/jhachmer/gomovie/internal/config"
)

func TestMain(m *testing.M) {
	config.Envs.OmdbApiKey = "TESTKEY"
	m.Run()
}

type OmdbMock struct {
}

func (r OmdbMock) SendRequest() (any, error) {
	return &Movie{}, nil
}

func (r OmdbMock) Validate() error {
	return nil
}

func Test_buildRequestURL(t *testing.T) {
	type args struct {
		r MediaRequest
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ID",
			args: args{r: MovieIDRequest{
				imdbID: "tt1234567",
			}},
			want:    "http://www.omdbapi.com/?apikey=TESTKEY&type=movie&i=tt1234567",
			wantErr: false,
		},
		{
			name: "Title and Year",
			args: args{r: MovieTitleRequest{
				title: "TestMovie",
				year:  "1984",
			}},
			want:    "http://www.omdbapi.com/?apikey=TESTKEY&type=movie&t=TestMovie&y=1984",
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
			got, err := buildRequestURL(tt.args.r)

			// Check if error status matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("buildRequestURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If an error is expected, no need to check further
			if tt.wantErr {
				return
			}

			// Parse URLs to compare query parameters in a normalized way
			gotURL, err := url.Parse(got)
			if err != nil {
				t.Fatalf("Failed to parse got URL: %v", err)
			}
			wantURL, err := url.Parse(tt.want)
			if err != nil {
				t.Fatalf("Failed to parse want URL: %v", err)
			}

			// Compare query parameters independently of order
			gotQuery, _ := url.ParseQuery(gotURL.RawQuery)
			wantQuery, _ := url.ParseQuery(wantURL.RawQuery)

			if !reflect.DeepEqual(gotQuery, wantQuery) {
				t.Errorf("buildRequestURL() query mismatch\ngot:  %v\nwant: %v", gotQuery, wantQuery)
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
			r := MovieTitleRequest{
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
			r := MovieIDRequest{
				imdbID: tt.fields.imdbID,
			}
			if err := r.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
