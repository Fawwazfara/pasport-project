package file_storage

type SavedFile struct {
	Name        string
	Size        int64
	ContentType string
}

type FileStorage interface {
	Save(nama string, isi []byte, contentType string) (SavedFile, error)
}
