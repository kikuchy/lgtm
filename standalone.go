// +build !appengine,!appenginevm

package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func createMux() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())

	e.Static("/", "public")

	return e
}

func loadConfig() config {
	return config{
		// TODO: 環境変数から読むなりファイルから読むなり
		os.Getenv("ADMIN_NAME"),
		os.Getenv("ADMIN_PASS"),
	}
}

func main() {
	e.Logger.Fatal(e.Start(":8080"))
}