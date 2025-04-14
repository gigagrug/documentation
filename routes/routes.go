package routes

import (
	"context"
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
}
type Document struct {
	ID      string
	Name    string
	Content string
	UserID  string
	Email   string
}

func (db *DB) Home(w http.ResponseWriter, r *http.Request) {
	var document Document
	rows, err := db.Query(context.Background(), `
		SELECT d.id, d.name, u.email
		FROM "Document" d 
		JOIN "User" u ON d.user_id = u.id
		ORDER BY d.updated_at DESC`)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "error getting posts", http.StatusInternalServerError)
	}
	defer rows.Close()

	documents := []Document{}
	for rows.Next() {
		err := rows.Scan(&document.ID, &document.Name, &document.Email)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "error getting posts", http.StatusInternalServerError)
			return
		}
		documents = append(documents, document)
	}
	auth := false
	user_id := r.Context().Value("user_id").(string)
	if user_id != "" {
		auth = true
	}

	data := struct {
		Documents []Document
		Auth      bool
		User      string
	}{
		Documents: documents,
		Auth:      auth,
		User:      user_id,
	}
	t := template.Must(template.ParseFiles("./frontend/frags/shell.html", "./frontend/index.html"))
	if err := t.Execute(w, data); err != nil {
		slog.Error(err.Error())
		http.Error(w, "error getting private", http.StatusInternalServerError)
		return
	}
	slog.Info("ok")
}

func (db *DB) Register(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./frontend/frags/shell.html", "./frontend/register.html"))
	if err := t.Execute(w, nil); err != nil {
		slog.Error(err.Error())
		http.Error(w, "error getting register", http.StatusInternalServerError)
		return
	}
}

func (db *DB) Login(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./frontend/frags/shell.html", "./frontend/login.html"))
	if err := t.Execute(w, nil); err != nil {
		slog.Error(err.Error())
		http.Error(w, "error getting login", http.StatusInternalServerError)
		return
	}
}

func (db *DB) Profile(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(string)
	var email string

	_ = db.QueryRow(context.Background(), `SELECT email FROM "User" WHERE id = $1`, user_id).Scan(&email)

	auth := false

	if user_id != "" {
		auth = true
	}
	data := struct {
		Auth  bool
		Email string
		ID    string
		User  string
	}{
		Auth:  auth,
		Email: email,
		ID:    user_id,
		User:  user_id,
	}
	t := template.Must(template.ParseFiles("./frontend/frags/shell.html", "./frontend/profile.html"))
	if err := t.Execute(w, data); err != nil {
		slog.Error(err.Error())
		http.Error(w, "error getting profile", http.StatusInternalServerError)
		return
	}
}

func (db *DB) Documents(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(string)
	userID := r.PathValue("userID")

	var document Document
	rows, err := db.Query(context.Background(), `
		SELECT id, name, content 
		FROM "Document"
		WHERE user_id = $1
		ORDER BY updated_at DESC
		`, userID)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "error getting posts", http.StatusInternalServerError)
	}
	defer rows.Close()

	documents := []Document{}
	for rows.Next() {
		err := rows.Scan(&document.ID, &document.Name, &document.Content)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "error getting posts", http.StatusInternalServerError)
			return
		}
		documents = append(documents, document)
	}
	auth := false
	if user_id != "" {
		auth = true
	}

	data := struct {
		Documents []Document
		Auth      bool
		User      string
	}{
		Documents: documents,
		Auth:      auth,
		User:      user_id,
	}
	t := template.Must(template.ParseFiles("./frontend/frags/shell.html", "./frontend/docs.html"))
	if err := t.Execute(w, data); err != nil {
		slog.Error(err.Error())
		http.Error(w, "error getting private", http.StatusInternalServerError)
		return
	}
	slog.Info("ok")
}

func (db *DB) Document(w http.ResponseWriter, r *http.Request) {
	docID := r.PathValue("docID")
	user_id := r.Context().Value("user_id").(string)

	var document Document
	err := db.QueryRow(context.Background(), `SELECT id, name, content, user_id FROM "Document" WHERE id = $1`, docID).Scan(&document.ID, &document.Name, &document.Content, &document.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Document not found", http.StatusNotFound)
		} else {
			slog.Error("Error fetching document: "+err.Error(), slog.Any("err", err))
			http.Error(w, "Error fetching document", http.StatusInternalServerError)
		}
		return
	}

	editAllowed := false
	if user_id == document.UserID {
		editAllowed = true
	}

	auth := false
	if user_id != "" {
		auth = true
	}

	data := struct {
		Document    Document
		Auth        bool
		EditAllowed bool
		User        string
	}{
		Document:    document,
		Auth:        auth,
		EditAllowed: editAllowed,
		User:        user_id,
	}

	t := template.Must(template.ParseFiles("./frontend/frags/shell.html", "./frontend/doc.html"))
	if err := t.Execute(w, data); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error rendering document", http.StatusInternalServerError)
		return
	}

}
