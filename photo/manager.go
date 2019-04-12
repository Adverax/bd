package photo

import (
	"io"
)

type Manager interface {
	Append(basename string, file io.Reader) error
	Delete(id int) error
	FindAll() ([]*File, error)
}

type Engine struct {
	Collector Collector
	Files     FileManager
}

func (engine *Engine) Append(basename string, file io.Reader) error {
	err := engine.Files.Append(basename, file)
	if err != nil {
		return err
	}

	err = engine.Collector.Append(basename)
	if err != nil {
		_ = engine.Files.Delete(basename)
		return err
	}

	return nil
}

func (engine *Engine) Delete(id int) error {
	row, err := engine.Collector.Find(id)
	if err != nil {
		return err
	}

	err = engine.Collector.Delete(id)
	if err != nil {
		return err
	}

	return engine.Files.Delete(row.File)
}

func (engine *Engine) FindAll() ([]*File, error) {
	return engine.Collector.FindAll()
}
