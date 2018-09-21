package main

import (
	"errors"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/microcosm-cc/bluemonday"
	"google.golang.org/appengine/datastore"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	p = bluemonday.UGCPolicy()
)

type (
	Entity struct {
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt time.Time
		IsDeleted bool
	}
	Image struct {
		Entity
		Url       string
		ServerKey *datastore.Key
	}
	Server struct {
		Entity
		Url string
	}
	Repository interface {
		SaveImage(image *Image) error
		LoadImages(count uint, page uint) ([]Image, error)
		RandomImageId() (int64, error)
		FindImageById(id int64) (*Image, error)
	}
	RepositoryGeneratorFunc func(c echo.Context) Repository

	Template struct {
		templates *template.Template
	}
)

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func init() {
	t := &Template{
		templates: template.Must(template.ParseGlob("template/*.html")),
	}
	e.Renderer = t

	// hook into the echo instance to create an endpoint group
	// and add specific middleware to it plus handlers
	e.Static("/", "public")
	e.GET("/browse", showAllImagesGenerator(RepositoryGenerator))
	e.GET("/g", showRandomImageGenerator(RepositoryGenerator))

	g := e.Group("/images")
	g.Use(middleware.CORS())
	g.POST("/", saveImageGenerator(RepositoryGenerator))
	g.GET("/:id", showImageGenerator(RepositoryGenerator))
	g.Static("/", "public/images")

	g = e.Group("/console")
	g.Use(middleware.BasicAuth(func(username string, password string, c echo.Context) (bool, error) {
		if username == "" || password == "" {
			return false, nil
		}
		u := cfg.adminName
		p := cfg.adminPass
		if u == "" || p == "" {
			return false, errors.New("username or password is not set")
		}
		if username == u && password == p {
			return true, nil
		}
		return false, nil
	}))
}

func showAllImagesGenerator(rg RepositoryGeneratorFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := rg(c)
		rawPage := c.QueryParam("rawPage")
		if rawPage == "" {
			rawPage = "1"
		}
		page, err := strconv.Atoi(rawPage)
		if err != nil {
			return err
		}
		images, err := r.LoadImages(50, uint(page))
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, "browse.template.html", map[string]interface{}{
			"prev":   page - 1,
			"next":   page + 1,
			"images": images,
		})
	}
}

func showRandomImageGenerator(rg RepositoryGeneratorFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := rg(c)
		id, err := r.RandomImageId()
		if err != nil {
			return err
		}
		if id < 0 {
			return errors.New("no images in DB")
		}
		return c.Redirect(http.StatusFound, "/images/"+strconv.Itoa(int(id)))
	}
}

func saveImageGenerator(rg RepositoryGeneratorFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := rg(c)
		println("Called")
		url := p.Sanitize(c.Request().FormValue("image_url"))
		image := &Image{
			Entity{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				IsDeleted: false,
			},
			url,
			nil,
		}
		if err := r.SaveImage(image); err != nil {
			return err
		}
		return c.HTML(http.StatusOK, "<p>done</p>")
	}
}

func showImageGenerator(rg RepositoryGeneratorFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := rg(c)
		rawId := c.Param("id")
		id, err := strconv.Atoi(rawId)
		if err != nil {
			return err
		}
		image, err := r.FindImageById(int64(id))
		return c.Render(http.StatusOK, "image.detail.template.html", image)
	}
}


