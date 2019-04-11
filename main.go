package main

import (
	"github.com/adverax/echo"
	"github.com/adverax/echo/database/sql"
	"github.com/adverax/middleware"
	"github.com/nfnt/resize"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

type File struct {
	Id   int    `json:"id"`
	File string `json:"file"`
}

// Filer is file manager
type Filer struct {
	Images     string // Path to the images folder
	Thumbnails string // Path to the thumbnails folder
	Width      uint   // Thumbnail width
	Height     uint   // Thumbnail height
}

// Append new image with required basename
func (filer *Filer) Append(basename string, file io.Reader) error {
	err := filer.makeFile(basename, file)
	if err != nil {
		return err
	}

	err = filer.makeThumbnail(
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
func (filer *Filer) Delete(basename string) error {
	_ = os.Remove(filer.Images + basename)
	_ = os.Remove(filer.Thumbnails + basename)
	return nil
}

// Store file into the images folder.
func (filer *Filer) makeFile(basename string, file io.Reader) error {
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

// Make thumbnail from src image.
func (filer *Filer) makeThumbnail(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	m := resize.Resize(filer.Width, filer.Height, img, resize.Lanczos3)

	out, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	return jpeg.Encode(out, m, nil)
}

type Manager struct {
	db sql.DB
}

// Find single File by identifier
func (mngr *Manager) Find(id int) (*File, error) {
	const query = "SELECT id, name FROM photo WHERE id = ?"
	row := new(File)
	err := mngr.db.QueryRow(query, id).Scan(&row.Id, &row.File)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// Append new row
func (mngr *Manager) Append(basename string) error {
	const query = "INSERT INTO photo SET file = ?"
	_, err := mngr.db.Exec(query, basename)
	return err
}

// Get list of files sorted by name.
func (mngr *Manager) FindAll() ([]*File, error) {
	const query = "SELECT id, name FROM photo ORDER BY file"
	rows, err := mngr.db.Query(query)
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

// Delete single row by identifier
func (mngr *Manager) Delete(id int) error {
	const query = "DELETE FROM photo WHERE id = ?"
	_, err := mngr.db.Exec(query, id)
	return err
}

// Http handler for get list of files.
func actionList(
	manager *Manager,
) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		list, err := manager.FindAll()
		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, list)
	}
}

// Http handler for delete single file.
func actionDelete(
	manager *Manager,
	filer *Filer,
) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.ParamInt("id", 0)
		row, err := manager.Find(id)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}

			return ctx.JSON(http.StatusOK, false)
		}

		err = manager.Delete(id)
		if err != nil {
			return err
		}

		err = filer.Delete(row.File)
		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, true)
	}
}

// Http handler for upload file.
func actionUpload(
	manager *Manager,
	filer *Filer,
) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		_, err := ctx.MultipartForm()
		if err != nil {
			return err
		}

		file, handler, err := ctx.Request().FormFile("file")
		if err != nil {
			return err
		}
		defer file.Close()

		filename := filepath.Base(handler.Filename)
		ext := path.Ext(filename)
		if ext != "jpeg" {
			return ctx.JSON(http.StatusBadRequest, false)
		}
		err = filer.Append(filename, file)
		if err != nil {
			return err
		}

		err = manager.Append(filename)
		if err != nil {
			_ = filer.Delete(filename)
			return err
		}

		return ctx.JSON(http.StatusOK, true)
	}
}

func main() {
	// Open database
	dsc := sql.DSC{
		Driver: "mysql",
		DSN: []*sql.DSN{
			{
				Host:     "127.0.0.1",
				Database: "mydatabase",
				Username: "username",
				Password: "password",
			},
		},
	}

	db, err := dsc.Open(nil)
	if err != nil {
		panic(err)
	}

	// Create managers
	workdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	manager := &Manager{
		db: db,
	}
	filer := &Filer{
		Images:     workdir + "/static/images/",
		Thumbnails: workdir + "/static/thumbnails/",
		Width:      64,
		Height:     64,
	}

	// Configure router
	e := echo.New()
	router := e.Router()

	router.Get(
		"/list",
		actionList(
			manager,
		),
	)

	router.Post(
		"/delete/{id}",
		actionDelete(
			manager,
			filer,
		),
	)

	router.Post(
		"/upload",
		actionUpload(
			manager,
			filer,
		),
	)

	// Creating handler for static content
	router.Use(middleware.Static("/static"))

	// Starting server
	log.Fatal(e.Start(":80"))
}
