package api

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Document struct {
	ID   string
	Name string
}

func (db *DB) CreateDocument(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	userID := r.Context().Value("user_id").(string)
	id, _ := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 15)

	if _, err := db.Exec(context.Background(), `
		INSERT INTO "Document" (id, name, user_id, updated_at) VALUES ($1, $2, $3, $4)`,
		id, name, userID, time.Now()); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error creating Document", http.StatusInternalServerError)
		return
	}

	document := Document{
		ID:   id,
		Name: name,
	}

	t := template.Must(template.ParseFiles("./frontend/frags/card.html"))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := t.Execute(w, document); err != nil {
		slog.Error("Error executing template: ")
		http.Error(w, "Error rendering document card", http.StatusInternalServerError)
	}
	slog.Info("CreateDoc: ok")
}

func (db *DB) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	docID := r.PathValue("docID")
	userID := r.Context().Value("user_id")
	name := r.FormValue("title")
	content := r.FormValue("content")

	var user_id string
	_ = db.QueryRow(context.Background(), `SELECT user_id FROM "Document" WHERE id = $1`, docID).Scan(&user_id)

	if user_id != userID {
		slog.Error("not auth")
		http.Error(w, "Not auth", http.StatusUnauthorized)
		return
	} else {
		if _, err := db.Exec(context.Background(), `UPDATE "Document" SET name = $1, content = $2, updated_at = $3 WHERE id = $4`, name, content, time.Now(), docID); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Error updating password", http.StatusInternalServerError)
			return
		}
	}

	slog.Info("UpdateDoc: ok")
}

func (db *DB) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	docID := r.PathValue("docID")
	userID := r.Context().Value("user_id")

	if _, err := db.Exec(context.Background(), `DELETE FROM "Document" WHERE id = $1 AND user_id = $2`, docID, userID); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error creating Document", http.StatusInternalServerError)
		return
	}
	slog.Info("DeleteDoc: ok")
}
