package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    e.GET("/:num", func(c echo.Context) error {
        return c.String(http.StatusOK, c.Param("num"))
    })
    e.Logger.Fatal(e.Start(":3000"))
}

