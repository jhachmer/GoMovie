package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type Example struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		want        any
		expectedErr error
	}{
		{
			name:        "Valid JSON",
			input:       Example{Name: "test", Value: 42},
			want:        Example{Name: "test", Value: 42},
			expectedErr: nil,
		},
		{
			name:        "Invalid JSON",
			input:       "invalid json",
			want:        Example{},
			expectedErr: errors.New("json decode err:"),
		},
		{
			name:        "Empty body",
			input:       nil,
			want:        Example{},
			expectedErr: errors.New("json decode err:"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.input != nil {
				var err error
				body, err = json.Marshal(tt.input)
				if err != nil {
					t.Fatalf("failed to marshal input: %v", err)
				}
			}
			response := httptest.NewRecorder()
			response.Body = bytes.NewBuffer(body)

			var got Example
			got, err := Decode[Example](response.Result())

			if (err != nil) != (tt.expectedErr != nil) {
				t.Fatalf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if err != nil && tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				if !strings.HasPrefix(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error prefix: %q, got: %q", tt.expectedErr.Error(), err.Error())
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected: %+v, got: %+v", tt.want, got)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		status      int
		want        string
		expectedErr error
	}{
		{
			name:        "Valid JSON",
			input:       Example{Name: "test", Value: 42},
			status:      http.StatusOK,
			want:        `{"name":"test","value":42}` + "\n",
			expectedErr: nil,
		},
		{
			name:        "Empty struct",
			input:       Example{},
			status:      http.StatusOK,
			want:        `{"name":"","value":0}` + "\n",
			expectedErr: nil,
		},
		{
			name:        "Invalid input (unsupported type)",
			input:       func() {},
			status:      http.StatusInternalServerError,
			want:        "",
			expectedErr: errors.New("json encode err:"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			err := Encode(w, r, tt.status, tt.input)

			if (err != nil) != (tt.expectedErr != nil) {
				t.Fatalf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if err != nil && tt.expectedErr != nil && !strings.HasPrefix(err.Error(), tt.expectedErr.Error()) {
				t.Fatalf("expected error prefix: %q, got: %q", tt.expectedErr.Error(), err.Error())
			}

			if status := w.Result().StatusCode; status != tt.status {
				t.Errorf("expected status code: %d, got: %d", tt.status, status)
			}

			if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("expected Content-Type: application/json, got: %s", contentType)
			}

			got := w.Body.String()
			if got != tt.want {
				t.Errorf("expected body: %q, got: %q", tt.want, got)
			}
		})
	}
}
