package photo

import (
	"github.com/adverax/echo/database/sql"
)

type File struct {
	Id   int    `json:"id"`
	File string `json:"file"`
}

type Collector interface {
	Find(id int) (*File, error)
	FindAll() ([]*File, error)
	Append(basename string) error
	Delete(id int) error
}

type CollectorEngine struct {
	DB sql.DB
}

// Find single File by identifier
func (c *CollectorEngine) Find(id int) (*File, error) {
	const query = "SELECT id, name FROM photo WHERE id = ?"
	row := new(File)
	err := c.DB.QueryRow(query, id).Scan(&row.Id, &row.File)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// Get list of files sorted by name.
func (c *CollectorEngine) FindAll() ([]*File, error) {
	const query = "SELECT id, name FROM photo ORDER BY file"
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*File
	for rows.Next() {
		row := new(File)
		err := rows.Scan(&row.Id, &row.File)
		if err != nil {
			return nil, err
		}
		res = append(res, row)
	}

	if err := rows.Err(); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return res, nil
}

// Append new row
func (c *CollectorEngine) Append(basename string) error {
	const query = "INSERT INTO photo SET file = ?"
	_, err := c.DB.Exec(query, basename)
	return err
}

// Delete single row by identifier
func (c *CollectorEngine) Delete(id int) error {
	const query = "DELETE FROM photo WHERE id = ?"
	_, err := c.DB.Exec(query, id)
	return err
}
