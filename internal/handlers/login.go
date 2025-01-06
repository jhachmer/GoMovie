package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jhachmer/gotocollection/internal/auth"
	"github.com/jhachmer/gotocollection/internal/types"
)

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}

func (h *Handler) CheckLoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusInternalServerError)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	var ok bool
	ok, err = h.store.CheckCredentials(username, password)
	if err != nil {
		//http.Error(w, "error while validating user", http.StatusInternalServerError)
		data := types.LoginData{Error: fmt.Errorf("error while logging in: %w", err)}
		renderTemplate(w, "index", data)
		return
	}
	if !ok {
		data := types.LoginData{Error: fmt.Errorf("invalid credentials")}
		renderTemplate(w, "index", data)
		return
	}
	tokenString, err := auth.CreateToken(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:    "golist",
		Value:   tokenString,
		Path:    "/",
		Domain:  "localhost",
		Expires: time.Now().Add(1 * time.Hour),
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/overview", http.StatusSeeOther)
}

func (h *Handler) RegisterSiteHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "register", nil)
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusInternalServerError)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	err = h.store.CreateUser(username, password)
	if err != nil {
		data := types.LoginData{Error: fmt.Errorf("error creating user %w", err)}
		renderTemplate(w, "register", data)
		return
	}
	http.Redirect(w, r, "login", http.StatusSeeOther)
}
