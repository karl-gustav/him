package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const (
	himURL = "https://him.as/tommekalender/?eiendomId=aa1582e2-6d78-4109-b844-2d6c6292c9fe"
)

var loc *time.Location
var months = map[string]time.Month{
	"januar":    time.January,
	"februar":   time.February,
	"mars":      time.March,
	"april":     time.April,
	"mai":       time.May,
	"juni":      time.June,
	"juli":      time.July,
	"august":    time.August,
	"september": time.September,
	"oktober":   time.October,
	"november":  time.November,
	"desember":  time.December,
}

type HIM struct {
	GarbageType string    `json:"type"`
	NextDate    time.Time `json:"nextDate"`
}

func init() {
	var err error
	loc, err = time.LoadLocation("Europe/Oslo")
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", himHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func himHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")
	c := colly.NewCollector()
	var dates []HIM

	c.OnHTML(".tommekalender__next__content", func(e *colly.HTMLElement) {
		ts, err := parseTS(e.ChildText(".tommekalender__next__date"))
		if err != nil {
			panic(err)
		}
		dates = append(dates, HIM{
			e.ChildText(".tommekalender__next__heading"),
			ts,
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(himURL)

	err := json.NewEncoder(res).Encode(dates)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func getUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Couldn't make GET request to %s:\n%v", url, err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("None 200 response code %v from %s:\n%s", resp.StatusCode, url, body)
	}
	return body, nil
}

func parseTS(dateString string) (time.Time, error) {
	now := time.Now()
	parts := strings.Split(dateString, ". ")
	date, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}
	month, ok := months[parts[1]]
	if !ok {
		return time.Time{}, fmt.Errorf("could not find %s in months map", parts[1])
	}
	ts := time.Date(now.Year(), month, date, 0, 0, 0, 0, loc)
	if ts.Before(now) {
		ts = ts.AddDate(1, 0, 0)
	}
	return ts, nil
}
