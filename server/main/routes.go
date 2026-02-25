package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"server/storage"
	"strings"

	"server/models"
)

type Server struct {
	db *sql.DB
}

func newServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) initRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/user/create", s.handleCreateUser)
	mux.HandleFunc("/api/v1/user/get", s.handleGetUser)

	return mux
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method invalid", http.StatusMethodNotAllowed)
		return
	}

	var req models.PostUser
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Department = strings.TrimSpace(req.Department)

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Age <= 0 {
		http.Error(w, "age must be > 0", http.StatusBadRequest)
		return
	}
	if req.Department == "" {
		http.Error(w, "department is required", http.StatusBadRequest)
		return
	}

	created, err := storage.CreateUser(s.db, req)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method invalid", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	if idStr == "" {
		http.Error(w, "missing query param: id", http.StatusBadRequest)
		return
	}

	u, err := storage.GetUserByID(s.db, idStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "db error: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}
