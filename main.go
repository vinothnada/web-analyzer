package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

type AnalysisResult struct {
	HTMLVersion   string         `json:"htmlVersion"`
	Title         string         `json:"title"`
	Headings      map[string]int `json:"headings"`
	InternalLinks int            `json:"internalLinks"`
	ExternalLinks int            `json:"externalLinks"`
	HasLoginForm  bool           `json:"hasLoginForm"`
}

type RequestPayload struct {
	URL string `json:"url"`
}

var urlRegex = regexp.MustCompile(`^https?://`) // Basic URL validation

func analyzePage(targetURL string) (*AnalysisResult, error) {
	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("URL returned status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := &AnalysisResult{
		Headings: make(map[string]int),
	}

	// Extract Title
	result.Title = doc.Find("title").Text()

	// Count Headings
	for i := 1; i <= 6; i++ {
		hTag := fmt.Sprintf("h%d", i)
		result.Headings[hTag] = doc.Find(hTag).Length()
	}

	// Count Links
	internalLinks, externalLinks := 0, 0
	parsedURL, _ := url.Parse(targetURL)
	// var wg sync.WaitGroup

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		hrefParsed, err := url.Parse(href)
		if err != nil || hrefParsed.Host == "" || hrefParsed.Host == parsedURL.Host {
			internalLinks++
		} else {
			externalLinks++
		}
	})

	result.InternalLinks = internalLinks
	result.ExternalLinks = externalLinks

	// Check for login form
	result.HasLoginForm = doc.Find("input[type='password']").Length() > 0

	return result, nil
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	logrus.Info("Received analysis request")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if !urlRegex.MatchString(payload.URL) {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	result, err := analyzePage(payload.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Info("Starting Web Analyzer API...")

	http.HandleFunc("/api/analyze", analyzeHandler)
	server := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(server.ListenAndServe())
}
