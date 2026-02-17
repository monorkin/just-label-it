package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrPinnedKeyframe is returned when attempting to move or delete a pinned keyframe.
var ErrPinnedKeyframe = errors.New("cannot modify pinned keyframe")

// Keyframe represents a labeled point in time on a video or audio file.
type Keyframe struct {
	ID          int64
	MediaFileID int64
	TimestampMs int64
	Description string
	Pinned      bool
	Labels      []Label
}

// EnsurePinnedKeyframe creates the pinned 0:00 keyframe for a media file if it doesn't exist.
func (d *DB) EnsurePinnedKeyframe(mediaFileID int64) error {
	var count int
	err := d.conn.QueryRow(
		`SELECT COUNT(*) FROM keyframes WHERE media_file_id = ? AND pinned = 1`,
		mediaFileID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("checking pinned keyframe for media file %d: %w", mediaFileID, err)
	}

	if count > 0 {
		return nil
	}

	_, err = d.conn.Exec(
		`INSERT INTO keyframes (media_file_id, timestamp_ms, pinned) VALUES (?, 0, 1)`,
		mediaFileID,
	)
	if err != nil {
		return fmt.Errorf("creating pinned keyframe for media file %d: %w", mediaFileID, err)
	}
	return nil
}

// KeyframesForMediaFile returns all keyframes for a media file, ordered by timestamp.
// Each keyframe includes its labels.
func (d *DB) KeyframesForMediaFile(mediaFileID int64) ([]Keyframe, error) {
	rows, err := d.conn.Query(
		`SELECT id, media_file_id, timestamp_ms, description, pinned
		 FROM keyframes WHERE media_file_id = ?
		 ORDER BY timestamp_ms ASC`,
		mediaFileID,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching keyframes for media file %d: %w", mediaFileID, err)
	}
	defer rows.Close()

	var keyframes []Keyframe
	for rows.Next() {
		var kf Keyframe
		if err := rows.Scan(&kf.ID, &kf.MediaFileID, &kf.TimestampMs, &kf.Description, &kf.Pinned); err != nil {
			return nil, fmt.Errorf("scanning keyframe: %w", err)
		}
		keyframes = append(keyframes, kf)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load labels for each keyframe.
	for i := range keyframes {
		labels, err := d.LabelsForKeyframe(keyframes[i].ID)
		if err != nil {
			return nil, err
		}
		keyframes[i].Labels = labels
	}

	return keyframes, nil
}

// GetKeyframe returns a single keyframe by ID.
func (d *DB) GetKeyframe(id int64) (*Keyframe, error) {
	kf := &Keyframe{}
	err := d.conn.QueryRow(
		`SELECT id, media_file_id, timestamp_ms, description, pinned FROM keyframes WHERE id = ?`,
		id,
	).Scan(&kf.ID, &kf.MediaFileID, &kf.TimestampMs, &kf.Description, &kf.Pinned)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetching keyframe %d: %w", id, err)
	}

	labels, err := d.LabelsForKeyframe(kf.ID)
	if err != nil {
		return nil, err
	}
	kf.Labels = labels

	return kf, nil
}

// CreateKeyframe adds a new keyframe at the given timestamp.
func (d *DB) CreateKeyframe(mediaFileID, timestampMs int64) (*Keyframe, error) {
	result, err := d.conn.Exec(
		`INSERT INTO keyframes (media_file_id, timestamp_ms) VALUES (?, ?)`,
		mediaFileID, timestampMs,
	)
	if err != nil {
		return nil, fmt.Errorf("creating keyframe at %dms for media file %d: %w", timestampMs, mediaFileID, err)
	}

	id, _ := result.LastInsertId()
	return d.GetKeyframe(id)
}

// UpdateKeyframeTimestamp moves a keyframe to a new timestamp. Rejects pinned keyframes.
func (d *DB) UpdateKeyframeTimestamp(id, timestampMs int64) error {
	kf, err := d.GetKeyframe(id)
	if err != nil {
		return err
	}
	if kf == nil {
		return fmt.Errorf("keyframe %d not found", id)
	}
	if kf.Pinned {
		return ErrPinnedKeyframe
	}

	_, err = d.conn.Exec(`UPDATE keyframes SET timestamp_ms = ? WHERE id = ?`, timestampMs, id)
	if err != nil {
		return fmt.Errorf("updating timestamp for keyframe %d: %w", id, err)
	}
	return nil
}

// UpdateKeyframeDescription updates a keyframe's description.
func (d *DB) UpdateKeyframeDescription(id int64, description string) error {
	result, err := d.conn.Exec(
		`UPDATE keyframes SET description = ? WHERE id = ?`,
		description, id,
	)
	if err != nil {
		return fmt.Errorf("updating description for keyframe %d: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("keyframe %d not found", id)
	}
	return nil
}

// DeleteKeyframe removes a keyframe. Rejects pinned keyframes.
func (d *DB) DeleteKeyframe(id int64) error {
	kf, err := d.GetKeyframe(id)
	if err != nil {
		return err
	}
	if kf == nil {
		return fmt.Errorf("keyframe %d not found", id)
	}
	if kf.Pinned {
		return ErrPinnedKeyframe
	}

	_, err = d.conn.Exec(`DELETE FROM keyframes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting keyframe %d: %w", id, err)
	}
	return nil
}
