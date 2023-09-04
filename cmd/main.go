package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/go-rod/rod"
	"github.com/labstack/echo/v4"
)

var browser = rod.New()
var positiveWords = generateWordSet("../data/positive-words.txt")
var negativeWords = generateWordSet("../data/negative-words.txt")

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
	c := make(chan float64, len(links))
	for i := 0; i < len(links); i++ {
		link := links[i].MustProperty("href").String()

		// ignore youtube shorts
		if matched, _ := regexp.Match(`.*/shorts/.*`, []byte(link)); !matched {
			go visitVideo(browser, link, c)
		}
	}

    score := 0.0
    for i := 0; i < len(links); i++ {
        s, _ := <- c
        fmt.Println(s)
        score += s
    }

    return context.String(http.StatusOK, fmt.Sprintf("%f", score))
}

func visitVideo(browser *rod.Browser, link string, c chan float64) {
	page := browser.MustConnect().MustPage(link).MustWaitDOMStable()
	defer page.MustClose()

	page.MustElement("#button-shape").MustClick()
	page.MustElementR("yt-formatted-string", "Show transcript").MustClick()

	score := 0
	wordCount := 0
	transcript := page.MustWaitStable().MustElements("yt-formatted-string.ytd-transcript-segment-renderer")
	for i := 0; i < len(transcript); i++ {
		s, w := getSentimentScore(transcript[i].MustText())
		score += s
		wordCount += w
	}

	averageSentiment := float64(score) / float64(wordCount)

	c <- averageSentiment
}

func generateWordSet(path string) *hashset.Set {

	ex, _ := os.Executable()
	dir := filepath.Dir(ex)
	f, err := os.Open(filepath.Join(dir, path))

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	set := hashset.New()
	for scanner.Scan() {
		set.Add(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return set
}

func getSentimentScore(s string) (int, int) {
	words := strings.Split(s, " ")

	score := 0
	for i, word := range words {
		if positiveWords.Contains(word) {
			score += 1
		}

		if negativeWords.Contains(word) {
			score -= 1
			if i > 0 && positiveWords.Contains(words[i-1]) {
				score -= 1
			}
		}
	}

	return score, len(words)
}
