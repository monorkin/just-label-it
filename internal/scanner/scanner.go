package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// mediaExtensions maps file extensions to their media type.
var mediaExtensions = map[string]string{
	// Images
	".jpg":  "image",
	".jpeg": "image",
	".png":  "image",
	".gif":  "image",
	".bmp":  "image",
	".webp": "image",
	".svg":  "image",
	".tiff": "image",
	".tif":  "image",
	".avif": "image",

	// Video
	".mp4":  "video",
	".webm": "video",
	".mkv":  "video",
	".avi":  "video",
	".mov":  "video",
	".m4v":  "video",
	".ogv":  "video",

	// Audio
	".mp3":  "audio",
	".wav":  "audio",
	".ogg":  "audio",
	".flac": "audio",
	".aac":  "audio",
	".m4a":  "audio",
	".wma":  "audio",
	".opus": "audio",
}

// File represents a discovered media file.
type File struct {
	Path      string // Relative path from the scan root.
	MediaType string // "image", "video", or "audio".
}

// Scan walks a directory tree and returns all recognized media files,
// sorted by their path (filepath.WalkDir visits in lexical order).
func Scan(root string) ([]File, error) {
	var files []File

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		mediaType, ok := mediaExtensions[ext]
		if !ok {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		files = append(files, File{Path: rel, MediaType: mediaType})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
