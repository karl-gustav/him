package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/karl-gustav/runlogger"
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

var (
	loc *time.Location
	log *runlogger.Logger
)

func init() {
	if os.Getenv("K_SERVICE") != "" { // Check if running in cloud run
		log = runlogger.StructuredLogger()
	} else {
		log = runlogger.PlainLogger()
	}
	var err error
	loc, err = time.LoadLocation("Europe/Oslo")
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		dates, err := getPickUp(r.Context(), loc)
		if err != nil {
			log.Errorf("Failed to get pick up dates from storage: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = json.NewEncoder(w).Encode(dates)
		if err != nil {
			log.Errorf("Failed to marshal pick up dates: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/trigger", func(w http.ResponseWriter, r *http.Request) {
		dates := getGarbagePickupDates(himURL)
		if len(dates) == 0 {
			log.Warningf("Failed to get pick up dates.")
			http.Error(w, "Failed to get pick up dates.", http.StatusInternalServerError)
			return
		}
		err := storePickUp(r.Context(), dates)
		if err != nil {
			log.Error("Failed to store pick up dates", log.Field("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Debug("Serving http://localhost:" + port)
	log.Error(http.ListenAndServe(":"+port, nil))
}

func getGarbagePickupDates(URL string) []HIM {
	c := colly.NewCollector()
	var dates []HIM

	c.OnHTML(".tommekalender__next__content", func(e *colly.HTMLElement) {
		ts, err := parseTS(e.ChildText(".tommekalender__next__date"))
		if err != nil {
			log.Error("Failed to find \".tommekalender__next__date\"", log.Field("error", err))
			return
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
	if dateString == "I dag" {
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	}

	dayString, monthString, ok := strings.Cut(dateString, ". ")
	if !ok {
		return time.Time{}, fmt.Errorf("could not cut datestring `%s`", dateString)
	}

	day, err := strconv.Atoi(dayString)
	if err != nil {
		return time.Time{}, err
	}
	month, ok := months[monthString]
	if !ok {
		return time.Time{}, fmt.Errorf("could not find %s in months map", monthString)
	}
	ts := time.Date(now.Year(), month, day, 0, 0, 0, 0, loc)
	if ts.AddDate(0, 6, 0).Before(now) {
		ts = ts.AddDate(1, 0, 0)
	}
	return ts, nil
}
