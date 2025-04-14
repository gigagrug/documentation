package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	*pgxpool.Pool
}
type RegisterForm struct {
	Email     string
	Password  string
	Password2 string
}

func generateSessionToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	return token, nil
}
func (db *DB) Register(w http.ResponseWriter, r *http.Request) {
	var user RegisterForm
	id, _ := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 15)
	user.Email = r.PostFormValue("email")
	user.Password = r.PostFormValue("password")
	user.Password2 = r.PostFormValue("password2")

	if user.Password != user.Password2 {
		slog.Error("Password not match")
		http.Error(w, "Passwords didn't match", http.StatusInternalServerError)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 15)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error creating password", http.StatusBadRequest)
		return
	}

	if _, err := db.Exec(context.Background(), `
		INSERT INTO "User" (id, email, password, updated_at) VALUES ($1, $2, $3, $4)`,
		id, user.Email, hash, time.Now()); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Email already exists", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("Register: ok")
}

func (db *DB) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string
		Password string
	}
	type User struct {
		ID       string
		Email    string
		Password string
	}

	body.Email = r.PostFormValue("email")
	body.Password = r.PostFormValue("password")

	var user User
	if err := db.QueryRow(context.Background(), `
		SELECT id, email, password 
		FROM "User" 
		WHERE email = $1`, body.Email).Scan(&user.ID, &user.Email, &user.Password); err != nil {
		slog.Error(err.Error())
		http.Error(w, "wrong email or password", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		slog.Error(err.Error())
		http.Error(w, "wrong email or password", http.StatusInternalServerError)
		return
	}
	token, err := generateSessionToken()
	if err != nil {
		log.Fatal(err)
	}
	timeDay := time.Now().Add(time.Hour * 24 * 30)
	timeNow := time.Now()
	if _, err := db.Exec(context.Background(), `
		INSERT INTO "Session" (session_token, expires, user_id, created_at) VALUES ($1, $2, $3, $4)`,
		token, timeDay, user.ID, timeNow); err != nil {
		slog.Error(err.Error())
		http.Error(w, "error logging in", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Expires:  timeDay,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
	slog.Info("Login: ok")
}

func (db *DB) Logout(w http.ResponseWriter, r *http.Request) {
	tokenString, err := r.Cookie("Authorization")
	if err != nil {
		slog.Error(err.Error())
		return
	}
	_, err = db.Exec(context.Background(), `
	DELETE FROM "Session" WHERE session_token = $1`, tokenString.Value)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "error logging out", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
	slog.Info("Logout: ok")
}
func (db *DB) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var user RegisterForm
	user_id := r.PathValue("userID")
	user.Email = r.PostFormValue("email")
	user.Password = r.PostFormValue("password")
	user.Password2 = r.PostFormValue("password2")

	if user.Email != "" {
		if _, err := db.Exec(context.Background(), `UPDATE "User" SET email = $1, updated_at = $2 WHERE id = $3`, user.Email, time.Now(), user_id); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Error updating email", http.StatusInternalServerError)
			return
		}
	}

	if user.Password != "" && user.Password2 != "" {
		if user.Password != user.Password2 {
			slog.Error("Passwords do not match")
			http.Error(w, "Passwords didn't match", http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Error creating password", http.StatusBadRequest)
			return
		}

		if _, err := db.Exec(context.Background(), `UPDATE "User" SET password = $1, updated_at = $2 WHERE id = $3`, hash, time.Now(), user_id); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Error updating password", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	slog.Info("UpdateProfile: ok")
}
