package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Decode[T any](r *http.Response) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("json decode err: %w", err)
	}
	return v, nil
}

func Encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("json encode err: %w", err)
	}
	return nil
}
