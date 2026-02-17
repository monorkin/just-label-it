package db

import (
	"database/sql"
	"fmt"
	"time"
)

// MediaFile represents a scanned media file in the database.
type MediaFile struct {
	ID          int64
	Path        string
	MediaType   string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpsertMediaFile inserts a media file or ignores it if the path already exists.
func (d *DB) UpsertMediaFile(path, mediaType string) error {
	_, err := d.conn.Exec(
		`INSERT INTO media_files (path, media_type) VALUES (?, ?) ON CONFLICT (path) DO NOTHING`,
		path, mediaType,
	)
	if err != nil {
		return fmt.Errorf("upserting media file %q: %w", path, err)
	}
	return nil
}

// GetMediaFile returns a single media file by ID.
func (d *DB) GetMediaFile(id int64) (*MediaFile, error) {
	m := &MediaFile{}
	err := d.conn.QueryRow(
		`SELECT id, path, media_type, description, created_at, updated_at FROM media_files WHERE id = ?`,
		id,
	).Scan(&m.ID, &m.Path, &m.MediaType, &m.Description, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetching media file %d: %w", id, err)
	}
	return m, nil
}

// FirstMediaFile returns the first media file ordered alphabetically by path.
func (d *DB) FirstMediaFile() (*MediaFile, error) {
	m := &MediaFile{}
	err := d.conn.QueryRow(
		`SELECT id, path, media_type, description, created_at, updated_at FROM media_files ORDER BY path ASC LIMIT 1`,
	).Scan(&m.ID, &m.Path, &m.MediaType, &m.Description, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetching first media file: %w", err)
	}
	return m, nil
}

// NavigationInfo holds the previous and next file IDs for navigation.
type NavigationInfo struct {
	PrevID     int64
	NextID     int64
	Index      int
	TotalCount int
}

// GetNavigation returns navigation context for a given media file.
// Files are ordered alphabetically by path with wrap-around.
func (d *DB) GetNavigation(currentID int64) (*NavigationInfo, error) {
	// Get the current file's path for ordering context.
	var currentPath string
	err := d.conn.QueryRow(`SELECT path FROM media_files WHERE id = ?`, currentID).Scan(&currentPath)
	if err != nil {
		return nil, fmt.Errorf("fetching path for media file %d: %w", currentID, err)
	}

	nav := &NavigationInfo{}

	// Total count.
	if err := d.conn.QueryRow(`SELECT COUNT(*) FROM media_files`).Scan(&nav.TotalCount); err != nil {
		return nil, fmt.Errorf("counting media files: %w", err)
	}

	// 1-based index of current file in alphabetical order.
	if err := d.conn.QueryRow(
		`SELECT COUNT(*) FROM media_files WHERE path <= ?`, currentPath,
	).Scan(&nav.Index); err != nil {
		return nil, fmt.Errorf("computing index for media file %d: %w", currentID, err)
	}

	// Previous file: the one just before in alphabetical order, wrapping to last.
	err = d.conn.QueryRow(
		`SELECT id FROM media_files WHERE path < ? ORDER BY path DESC LIMIT 1`, currentPath,
	).Scan(&nav.PrevID)
	if err == sql.ErrNoRows {
		// Wrap to last file.
		d.conn.QueryRow(`SELECT id FROM media_files ORDER BY path DESC LIMIT 1`).Scan(&nav.PrevID)
	} else if err != nil {
		return nil, fmt.Errorf("fetching previous media file: %w", err)
	}

	// Next file: the one just after in alphabetical order, wrapping to first.
	err = d.conn.QueryRow(
		`SELECT id FROM media_files WHERE path > ? ORDER BY path ASC LIMIT 1`, currentPath,
	).Scan(&nav.NextID)
	if err == sql.ErrNoRows {
		// Wrap to first file.
		d.conn.QueryRow(`SELECT id FROM media_files ORDER BY path ASC LIMIT 1`).Scan(&nav.NextID)
	} else if err != nil {
		return nil, fmt.Errorf("fetching next media file: %w", err)
	}

	return nav, nil
}

// UpdateDescription updates a media file's description.
func (d *DB) UpdateDescription(id int64, description string) error {
	result, err := d.conn.Exec(
		`UPDATE media_files SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		description, id,
	)
	if err != nil {
		return fmt.Errorf("updating description for media file %d: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("media file %d not found", id)
	}
	return nil
}

// MediaFileCount returns the total number of media files.
func (d *DB) MediaFileCount() (int, error) {
	var count int
	if err := d.conn.QueryRow(`SELECT COUNT(*) FROM media_files`).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting media files: %w", err)
	}
	return count, nil
}
