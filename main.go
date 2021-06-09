package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blendle/zapdriver"
	"github.com/gocolly/colly"
)

const (
	himURL = "https://him.as/tommekalender/?eiendomId=aa1582e2-6d78-4109-b844-2d6c6292c9fe"
)

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
}

func main() {
	structuredLogger, _ := zapdriver.NewProduction()
	defer structuredLogger.Sync() // flushes buffer, if any
	logger := structuredLogger.Sugar()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		dates, err := getPickUp(r.Context())
		if err != nil {
			logger.Errorf("Failed to get pick up dates from storage: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = json.NewEncoder(w).Encode(dates)
		if err != nil {
			logger.Errorf("Failed to marshal pick up dates: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/trigger", func(w http.ResponseWriter, r *http.Request) {
		dates := getGarbagePickupDates(himURL)
		if len(dates) == 0 {
			logger.Errorf("Failed to get pick up dates.")
			http.Error(w, "Failed to get pick up dates.", http.StatusInternalServerError)
			return
		}
		err := storePickUp(r.Context(), dates)
		if err != nil {
			logger.Errorf("Failed to store pick up dates: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getGarbagePickupDates(URL string) []HIM {
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

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})

	c.Visit(URL)
	return dates
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
	ts := time.Date(now.Year(), month, date, 0, 0, 0, 0, time.UTC)
	if ts.AddDate(0, 6, 0).Before(now) {
		ts = ts.AddDate(1, 0, 0)
	}
	return ts, nil
}
