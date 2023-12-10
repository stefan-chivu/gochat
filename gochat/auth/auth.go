package auth

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/stefan-chivu/gochat/gochat/models"
)

var store *sessions.CookieStore

func NewCookieStore() error {
	key := make([]byte, 64)

	_, err := rand.Read(key)
	if err != nil {
		return err
	}

	store = sessions.NewCookieStore(key)

	return nil
}

func Secret(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Print secret message
	fmt.Fprintln(w, "The cake is a lie!")
}

func Login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	fmt.Printf("%v", r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Parse form failed", http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")

	// TODO better valid username check
	if username == "" {
		// error case
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	user := &models.User{Username: username}
	if authenticateUser(user) {
		// Set user as authenticated
		session.Values["authenticated"] = true
		session.Save(r, w)

		// redirect to whatever
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)
}

func authenticateUser(u *models.User) bool {
	// TODO: Auth user here

	return true
}
