package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/go-rod/rod"
	"github.com/labstack/echo/v4"
)

type Pair[T, U any] struct {
	First  T
	Second U
}

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
	score := 0.0
	visits := 0
	for i := 0; i < len(links); i++ {
		link := links[i].MustProperty("href").String()

		// ignore youtube shorts
		if matched, _ := regexp.Match(`.*/shorts/.*`, []byte(link)); !matched {
			increment := visitVideo(browser, link)
			score += increment
            // ignore insignificant stats
			if increment != 0 {
				visits++
			}

		}
	}

	return context.String(http.StatusOK, fmt.Sprintf("%f", score))
}

func visitVideo(browser *rod.Browser, link string) float64 {
	page := browser.MustConnect().MustPage(link)
	defer page.MustClose()

	page.MustElement("#button-shape").MustClick()
	hasTranscript, transcriptButton, err := page.HasR("yt-formatted-string", "Show transcript")

	if err != nil || !hasTranscript {
		return 0.0
	}

	transcriptButton.MustClick()

	transcript := page.MustWaitStable().MustElements("yt-formatted-string.ytd-transcript-segment-renderer")
	c := make(chan Pair[int, int], 2*len(transcript))
	for i := 0; i < len(transcript); i++ {
		go getSentimentScore(transcript[i].MustText(), c)
	}

	score := 0
	wordCount := 0
	for i := 0; i < len(transcript); i++ {
		res := <-c
		score += res.First
		wordCount += res.Second
	}

	if wordCount == 0 {
		return 0.0
	}

	return float64(score) / float64(wordCount)

}

func generateWordSet(path string) *hashset.Set {
	f, err := os.Open(path)

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

func getSentimentScore(s string, c chan Pair[int, int]) {
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
	c <- Pair[int, int]{score, len(words)}
}
