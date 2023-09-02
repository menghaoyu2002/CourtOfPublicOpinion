package main

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-rod/rod"
	"github.com/labstack/echo/v4"
)

var browser = rod.New()

func main() {
	e := echo.New()
	e.GET("/search", handleSearch)
	e.Logger.Fatal(e.Start(":3000"))
}

func handleSearch(context echo.Context) error {
	s := context.QueryParam("search_query")

	if s == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing Search String")
	}

	youtubeUrl := "https://www.youtube.com/results?search_query=%s"
	yt := fmt.Sprintf(youtubeUrl, s)

	page := browser.MustConnect().MustPage(yt)
	defer page.MustClose()

	links := page.MustWaitDOMStable().MustElements("a#video-title")
	for i := 0; i < len(links); i++ {
		link := links[i].MustProperty("href").String()

        // ignore youtube shorts
		if matched, _ := regexp.Match(`.*/shorts/.*`, []byte(link)); !matched {
			go visitVideo(browser, link)
		}
	}

	return context.String(http.StatusOK, yt)
}

func visitVideo(browser *rod.Browser, link string) {
	page := browser.MustConnect().MustPage(link)
	defer page.MustClose()

}
