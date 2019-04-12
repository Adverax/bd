package photo

import (
	"io"
	"os"
)

type FileManager interface {
	Append(basename string, file io.Reader) error
	Delete(basename string) error
}

type FileEngine struct {
	ThumbnailManager ThumbnailManager
	Images           string // Path to the images folder
	Thumbnails       string // Path to the thumbnails folder
}

// Append new image with required basename
func (filer *FileEngine) Append(basename string, file io.Reader) error {
	err := filer.makeFile(basename, file)
	if err != nil {
		return err
	}

	err = filer.ThumbnailManager.Execute(
		filer.Images+basename,
		filer.Thumbnails+basename,
	)

	if err != nil {
		_ = filer.Delete(basename)
		return err
	}

	return nil
}

// Delete image by basename
func (filer *FileEngine) Delete(basename string) error {
	_ = os.Remove(filer.Images + basename)
	_ = os.Remove(filer.Thumbnails + basename)
	return nil
}

// Store file into the images folder.
func (filer *FileEngine) makeFile(basename string, file io.Reader) error {
	fileName := filer.Images + basename
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}

	return nil
}
