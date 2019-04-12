package main

import (
	"bd/photo"
	"github.com/adverax/echo"
	"github.com/adverax/echo/database/sql"
	"github.com/adverax/echo/middleware"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

// Http handler for get list of files.
func actionList(
	manager photo.Manager,
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
	manager photo.Manager,
) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.ParamInt("id", 0)
		err := manager.Delete(id)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}

			return ctx.JSON(http.StatusOK, false)
		}

		return ctx.JSON(http.StatusOK, true)
	}
}

// Http handler for upload file.
func actionUpload(
	manager photo.Manager,
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

		basename := filepath.Base(handler.Filename)
		ext := path.Ext(basename)
		if ext != "jpeg" {
			return ctx.JSON(http.StatusBadRequest, false)
		}
		err = manager.Append(basename, file)
		if err != nil {
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

	workdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	manager := &photo.Engine{
		Collector: &photo.CollectorEngine{
			DB: db,
		},
		Files: &photo.FileEngine{
			Images:     workdir + "/static/images/",
			Thumbnails: workdir + "/static/thumbnails/",
			ThumbnailManager: &photo.ThumbnailEngine{
				Width:  64,
				Height: 64,
			},
		},
	}

	// Configure router
	e := echo.New()
	router := e.Router()

	// Creating handler for static content
	router.Use(middleware.Static("/static"))

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
		),
	)

	router.Post(
		"/upload",
		actionUpload(
			manager,
		),
	)

	// Starting server
	log.Fatal(e.Start(":80"))
}
