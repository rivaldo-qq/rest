package repository

// ImageRepository mendefinisikan kontrak untuk penyimpanan gambar
type ImageRepository interface {
	Upload(fileBytes []byte, filename string) (string, error) // Return URL gambar
	Delete(filepath string) error
}
