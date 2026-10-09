package file_storage

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

type localStorage struct {
	dir string
}

func NewLocalStorage(dir string) *localStorage {
	return &localStorage{dir: dir}
}

func (s *localStorage) Save(_ string, isi []byte, contentType string) (SavedFile, error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return SavedFile{}, err
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return SavedFile{}, err
	}
	storedName := hex.EncodeToString(buf) + extFor(contentType)

	if err := os.WriteFile(filepath.Join(s.dir, storedName), isi, 0o644); err != nil {
		return SavedFile{}, err
	}
	return SavedFile{
		Name:        storedName,
		Size:        int64(len(isi)),
		ContentType: contentType,
	}, nil
}

func extFor(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "application/pdf":
		return ".pdf"
	default:
		return ".bin"
	}
}
