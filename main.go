package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	dataDir    = "data"
	metaFile   = "data/meta.json"
	notesDir   = "data/notes"
	uploadsDir = "data/uploads"
)

func init() {
	// Create data and uploads directories if they don't exist
	for _, dir := range []string{dataDir, notesDir, uploadsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}
}

func main() {
	// API routes
	http.HandleFunc("/api/meta", handleMeta)
	http.HandleFunc("/api/notes/", handleNotes)
	http.HandleFunc("/api/images", handleImageUpload)

	// Serve static uploads
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// Serve application index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			http.ServeFile(w, r, "index.html")
			return
		}
		// Fallback to index.html for frontend routing (SPA support)
		http.ServeFile(w, r, "index.html")
	})

	port := "8342"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	log.Printf("Server listening on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func handleMeta(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if _, err := os.Stat(metaFile); os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("null"))
			return
		}
		http.ServeFile(w, r, metaFile)

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		if err := os.WriteFile(metaFile, body, 0644); err != nil {
			http.Error(w, "failed to save meta data", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleNotes(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "invalid note id", http.StatusBadRequest)
		return
	}

	notePath := filepath.Join(notesDir, id+".json")

	switch r.Method {
	case http.MethodGet:
		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("null"))
			return
		}
		http.ServeFile(w, r, notePath)

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		if err := os.WriteFile(notePath, body, 0644); err != nil {
			http.Error(w, "failed to save note data", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		if err := os.Remove(notePath); err != nil && !os.IsNotExist(err) {
			http.Error(w, "failed to delete note file", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleImageUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit to 10 MB form parses
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "failed to parse image form-file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate clean unique name: nano timestamp prefix
	cleanName := strings.ReplaceAll(handler.Filename, " ", "_")
	uniqueName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(cleanName))
	destPath := filepath.Join(uploadsDir, uniqueName)

	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "failed to create destination file on server", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		http.Error(w, "failed to write image to disk", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"url":      "/uploads/" + uniqueName,
		"filename": uniqueName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
