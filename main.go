package main

import (
	"context"
	"emanwel/api"
	"emanwel/routes"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func main() {
	conn, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	db = conn
	defer conn.Close()

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/assets/"))))

	routes := &routes.DB{Pool: db}
	mux.HandleFunc("GET /{$}", AuthCheck(routes.Home, db, false))
	mux.HandleFunc("GET /profile", AuthCheck(routes.Profile, db, true))
	mux.HandleFunc("GET /u/{userID}/", AuthCheck(routes.Documents, db, false))
	mux.HandleFunc("GET /doc/{docID}/{$}", AuthCheck(routes.Document, db, false))

	mux.HandleFunc("GET /login", routes.Login)
	mux.HandleFunc("GET /register", routes.Register)

	api := &api.DB{Pool: db}
	mux.HandleFunc("POST /api/createDocument/{$}", AuthCheck(api.CreateDocument, db, true))
	mux.HandleFunc("PUT /api/{docID}/updateDocument/{$}", AuthCheck(api.UpdateDocument, db, true))
	mux.HandleFunc("DELETE /api/{docID}/deleteDocument/{$}", AuthCheck(api.DeleteDocument, db, true))

	mux.HandleFunc("PATCH /user/{userID}/updateProfile/{$}", AuthCheck(api.UpdateProfile, db, true))
	mux.HandleFunc("POST /user/register/{$}", api.Register)
	mux.HandleFunc("POST /user/login/{$}", api.Login)
	mux.HandleFunc("POST /user/logout/{$}", api.Logout)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}

type Session struct {
	SessionToken string
	Expires      time.Time
	UserID       string
}

func AuthCheck(f http.HandlerFunc, db *pgxpool.Pool, authNeeded bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:     "Authorization",
			Value:    "",
			MaxAge:   -1,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}

		tokenString, err := r.Cookie("Authorization")
		if err != nil {
			if authNeeded {
				slog.Error("Cookie not found: " + err.Error())
				http.SetCookie(w, &cookie)
				http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", "")
			r = r.WithContext(ctx)
		} else {
			var session Session
			err := db.QueryRow(context.Background(), `
				SELECT session_token, expires, user_id
				FROM "Session"
				WHERE session_token = $1`, tokenString.Value).Scan(&session.SessionToken, &session.Expires, &session.UserID)
			if err != nil {
				if err == pgx.ErrNoRows {
					slog.Warn("Session not found for token: " + tokenString.Value)
				} else {
					slog.Error("Database query failed: " + err.Error())
				}
				http.SetCookie(w, &cookie)
				http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
				return
			}

			if session.Expires.Before(time.Now()) {
				http.SetCookie(w, &cookie)
				http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", session.UserID)
			r = r.WithContext(ctx)
		}
		f(w, r)
	}
}
