package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/monorkin/just-label-it/internal/db"
)

// viewerData is the template data for the viewer page.
type viewerData struct {
	File      *db.MediaFile
	Labels    []db.Label
	Keyframes []db.Keyframe
	Nav       *db.NavigationInfo
}

// handleIndex redirects to the first media file, or shows an empty state.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	first, err := s.db.FirstMediaFile()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error fetching first media file: %v", err)
		return
	}

	if first == nil {
		s.renderTemplate(w, "viewer.html", nil)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/files/%d", first.ID), http.StatusFound)
}

// handleViewFile renders the viewer page for a specific media file.
func (s *Server) handleViewFile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	file, err := s.db.GetMediaFile(id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error fetching media file %d: %v", id, err)
		return
	}
	if file == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	labels, err := s.db.LabelsForMediaFile(id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error fetching labels for media file %d: %v", id, err)
		return
	}

	nav, err := s.db.GetNavigation(id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error fetching navigation for media file %d: %v", id, err)
		return
	}

	var keyframes []db.Keyframe
	if file.MediaType == "video" || file.MediaType == "audio" {
		if err := s.db.EnsurePinnedKeyframe(id); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("error ensuring pinned keyframe for media file %d: %v", id, err)
			return
		}

		keyframes, err = s.db.KeyframesForMediaFile(id)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("error fetching keyframes for media file %d: %v", id, err)
			return
		}
	}

	s.renderTemplate(w, "viewer.html", viewerData{
		File:      file,
		Labels:    labels,
		Keyframes: keyframes,
		Nav:       nav,
	})
}

// handleServeMedia serves a media file from the filesystem.
// Path traversal is prevented by resolving to an absolute path and checking the prefix.
func (s *Server) handleServeMedia(w http.ResponseWriter, r *http.Request) {
	relPath := r.PathValue("path")
	if relPath == "" {
		http.Error(w, "Missing path", http.StatusBadRequest)
		return
	}

	absPath, err := filepath.Abs(filepath.Join(s.mediaRoot, relPath))
	if err != nil || !strings.HasPrefix(absPath, s.mediaRoot) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, absPath)
}

// handleAddFileLabel adds a label to a media file.
func (s *Server) handleAddFileLabel(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	label, err := s.db.FindOrCreateLabel(strings.TrimSpace(body.Name))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error creating label %q: %v", body.Name, err)
		return
	}

	if err := s.db.AddMediaLabel(id, label.ID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error adding label to media file %d: %v", id, err)
		return
	}

	respondJSON(w, http.StatusCreated, label)
}

// handleRemoveFileLabel removes a label from a media file.
func (s *Server) handleRemoveFileLabel(w http.ResponseWriter, r *http.Request) {
	fileID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	labelID, err := parseID(r, "lid")
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	if err := s.db.RemoveMediaLabel(fileID, labelID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error removing label %d from media file %d: %v", labelID, fileID, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateFileDescription updates a media file's description.
func (s *Server) handleUpdateFileDescription(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.db.UpdateDescription(id, body.Description); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error updating description for media file %d: %v", id, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCreateKeyframe adds a keyframe to a media file.
func (s *Server) handleCreateKeyframe(w http.ResponseWriter, r *http.Request) {
	fileID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	var body struct {
		TimestampMs int64 `json:"timestamp_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	kf, err := s.db.CreateKeyframe(fileID, body.TimestampMs)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error creating keyframe for media file %d: %v", fileID, err)
		return
	}

	respondJSON(w, http.StatusCreated, kf)
}

// handleUpdateKeyframe moves a keyframe to a new timestamp.
func (s *Server) handleUpdateKeyframe(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid keyframe ID", http.StatusBadRequest)
		return
	}

	var body struct {
		TimestampMs int64 `json:"timestamp_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.db.UpdateKeyframeTimestamp(id, body.TimestampMs); err != nil {
		if errors.Is(err, db.ErrPinnedKeyframe) {
			http.Error(w, "Cannot move pinned keyframe", http.StatusForbidden)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error updating keyframe %d: %v", id, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteKeyframe deletes a keyframe.
func (s *Server) handleDeleteKeyframe(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid keyframe ID", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteKeyframe(id); err != nil {
		if errors.Is(err, db.ErrPinnedKeyframe) {
			http.Error(w, "Cannot delete pinned keyframe", http.StatusForbidden)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error deleting keyframe %d: %v", id, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleAddKeyframeLabel adds a label to a keyframe.
func (s *Server) handleAddKeyframeLabel(w http.ResponseWriter, r *http.Request) {
	kfID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid keyframe ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	label, err := s.db.FindOrCreateLabel(strings.TrimSpace(body.Name))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error creating label %q: %v", body.Name, err)
		return
	}

	if err := s.db.AddKeyframeLabel(kfID, label.ID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error adding label to keyframe %d: %v", kfID, err)
		return
	}

	respondJSON(w, http.StatusCreated, label)
}

// handleRemoveKeyframeLabel removes a label from a keyframe.
func (s *Server) handleRemoveKeyframeLabel(w http.ResponseWriter, r *http.Request) {
	kfID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid keyframe ID", http.StatusBadRequest)
		return
	}

	labelID, err := parseID(r, "lid")
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	if err := s.db.RemoveKeyframeLabel(kfID, labelID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error removing label %d from keyframe %d: %v", labelID, kfID, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateKeyframeDescription updates a keyframe's description.
func (s *Server) handleUpdateKeyframeDescription(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "Invalid keyframe ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.db.UpdateKeyframeDescription(id, body.Description); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error updating description for keyframe %d: %v", id, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleSearchLabels returns labels matching a query string.
func (s *Server) handleSearchLabels(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondJSON(w, http.StatusOK, []db.Label{})
		return
	}

	labels, err := s.db.SearchLabels(query)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error searching labels for %q: %v", query, err)
		return
	}

	if labels == nil {
		labels = []db.Label{}
	}

	respondJSON(w, http.StatusOK, labels)
}

// renderTemplate executes a named template with the given data.
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("error rendering template %q: %v", name, err)
	}
}

// parseID extracts an int64 path parameter from the request.
func parseID(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(r.PathValue(name), 10, 64)
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
