package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID   int    `json:"UserID"`
	Username string `json:"Username"`
	Active   int    `json:"Active"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	passwordHash, err := h.store.AdminLoginQuery(creds.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.store.GetUsers()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserID, &user.Username, &user.Active); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) ToggleActiveHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID int `json:"userId"`
		Active int `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		fmt.Print(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.store.ToggleUserActive(request.Active, request.UserID)
	if err != nil {
		fmt.Print(err)
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AdminHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin", nil)
}
