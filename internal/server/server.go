package server

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/monorkin/just-label-it/internal/db"
	"github.com/monorkin/just-label-it/web"
)

// Server holds the dependencies for all HTTP handlers.
type Server struct {
	db        *db.DB
	templates *template.Template
	mediaRoot string
}

// New creates a Server and returns a configured http.Handler.
func New(database *db.DB, mediaRoot string) (http.Handler, error) {
	absRoot, err := filepath.Abs(mediaRoot)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("").Funcs(templateFuncs()).ParseFS(web.Templates, "templates/*.html")
	if err != nil {
		return nil, err
	}

	s := &Server{
		db:        database,
		templates: tmpl,
		mediaRoot: absRoot,
	}

	mux := http.NewServeMux()
	s.routes(mux)

	// Wrap with CSRF protection.
	csrf := http.NewCrossOriginProtection()
	handler := csrf.Handler(mux)

	return handler, nil
}

func (s *Server) routes(mux *http.ServeMux) {
	// Static assets (embedded).
	staticFS, err := fs.Sub(web.Static, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Pages.
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /files/{id}", s.handleViewFile)

	// Media file serving.
	mux.HandleFunc("GET /media/{path...}", s.handleServeMedia)

	// File labels.
	mux.HandleFunc("POST /files/{id}/labels", s.handleAddFileLabel)
	mux.HandleFunc("DELETE /files/{id}/labels/{lid}", s.handleRemoveFileLabel)

	// File description.
	mux.HandleFunc("PUT /files/{id}/description", s.handleUpdateFileDescription)

	// Keyframes.
	mux.HandleFunc("POST /files/{id}/keyframes", s.handleCreateKeyframe)
	mux.HandleFunc("PUT /keyframes/{id}", s.handleUpdateKeyframe)
	mux.HandleFunc("DELETE /keyframes/{id}", s.handleDeleteKeyframe)

	// Keyframe labels.
	mux.HandleFunc("POST /keyframes/{id}/labels", s.handleAddKeyframeLabel)
	mux.HandleFunc("DELETE /keyframes/{id}/labels/{lid}", s.handleRemoveKeyframeLabel)

	// Keyframe description.
	mux.HandleFunc("PUT /keyframes/{id}/description", s.handleUpdateKeyframeDescription)

	// Label search API.
	mux.HandleFunc("GET /api/labels", s.handleSearchLabels)
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"isVideo": func(mediaType string) bool { return mediaType == "video" },
		"isAudio": func(mediaType string) bool { return mediaType == "audio" },
		"isImage": func(mediaType string) bool { return mediaType == "image" },
		"isTemporal": func(mediaType string) bool {
			return mediaType == "video" || mediaType == "audio"
		},
	}
}
