package db

import (
	"fmt"
)

// Label represents a reusable tag.
type Label struct {
	ID   int64
	Name string
}

// FindOrCreateLabel returns an existing label by name, or creates one.
func (d *DB) FindOrCreateLabel(name string) (*Label, error) {
	_, err := d.conn.Exec(`INSERT INTO labels (name) VALUES (?) ON CONFLICT (name) DO NOTHING`, name)
	if err != nil {
		return nil, fmt.Errorf("creating label %q: %w", name, err)
	}

	l := &Label{}
	err = d.conn.QueryRow(`SELECT id, name FROM labels WHERE name = ?`, name).Scan(&l.ID, &l.Name)
	if err != nil {
		return nil, fmt.Errorf("fetching label %q: %w", name, err)
	}
	return l, nil
}

// SearchLabels returns labels matching a prefix query, limited to 10 results.
func (d *DB) SearchLabels(query string) ([]Label, error) {
	rows, err := d.conn.Query(
		`SELECT id, name FROM labels WHERE name LIKE ? ORDER BY name ASC LIMIT 10`,
		query+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("searching labels for %q: %w", query, err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var l Label
		if err := rows.Scan(&l.ID, &l.Name); err != nil {
			return nil, fmt.Errorf("scanning label: %w", err)
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// AddMediaLabel associates a label with a media file.
func (d *DB) AddMediaLabel(mediaFileID, labelID int64) error {
	_, err := d.conn.Exec(
		`INSERT INTO media_labels (media_file_id, label_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		mediaFileID, labelID,
	)
	if err != nil {
		return fmt.Errorf("adding label %d to media file %d: %w", labelID, mediaFileID, err)
	}
	return nil
}

// RemoveMediaLabel removes a label association from a media file.
func (d *DB) RemoveMediaLabel(mediaFileID, labelID int64) error {
	_, err := d.conn.Exec(
		`DELETE FROM media_labels WHERE media_file_id = ? AND label_id = ?`,
		mediaFileID, labelID,
	)
	if err != nil {
		return fmt.Errorf("removing label %d from media file %d: %w", labelID, mediaFileID, err)
	}
	return nil
}

// LabelsForMediaFile returns all labels attached to a media file.
func (d *DB) LabelsForMediaFile(mediaFileID int64) ([]Label, error) {
	rows, err := d.conn.Query(
		`SELECT l.id, l.name FROM labels l
		 JOIN media_labels ml ON ml.label_id = l.id
		 WHERE ml.media_file_id = ?
		 ORDER BY l.name ASC`,
		mediaFileID,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching labels for media file %d: %w", mediaFileID, err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var l Label
		if err := rows.Scan(&l.ID, &l.Name); err != nil {
			return nil, fmt.Errorf("scanning label: %w", err)
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// AddKeyframeLabel associates a label with a keyframe.
func (d *DB) AddKeyframeLabel(keyframeID, labelID int64) error {
	_, err := d.conn.Exec(
		`INSERT INTO keyframe_labels (keyframe_id, label_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		keyframeID, labelID,
	)
	if err != nil {
		return fmt.Errorf("adding label %d to keyframe %d: %w", labelID, keyframeID, err)
	}
	return nil
}

// RemoveKeyframeLabel removes a label association from a keyframe.
func (d *DB) RemoveKeyframeLabel(keyframeID, labelID int64) error {
	_, err := d.conn.Exec(
		`DELETE FROM keyframe_labels WHERE keyframe_id = ? AND label_id = ?`,
		keyframeID, labelID,
	)
	if err != nil {
		return fmt.Errorf("removing label %d from keyframe %d: %w", labelID, keyframeID, err)
	}
	return nil
}

// LabelsForKeyframe returns all labels attached to a keyframe.
func (d *DB) LabelsForKeyframe(keyframeID int64) ([]Label, error) {
	rows, err := d.conn.Query(
		`SELECT l.id, l.name FROM labels l
		 JOIN keyframe_labels kl ON kl.label_id = l.id
		 WHERE kl.keyframe_id = ?
		 ORDER BY l.name ASC`,
		keyframeID,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching labels for keyframe %d: %w", keyframeID, err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var l Label
		if err := rows.Scan(&l.ID, &l.Name); err != nil {
			return nil, fmt.Errorf("scanning label: %w", err)
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}
