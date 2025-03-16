package analyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/vinothnada/web-analyzer/internal/types"
)

var urlRegex = regexp.MustCompile(`^https?://`)

func GetResults() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		logrus.Info("Received analysis request")
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var payload types.RequestPayload
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

		json.NewEncoder(w).Encode(result)
	}
}

func analyzePage(targetURL string) (*types.AnalyzeResultes, error) {
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

	result := &types.AnalyzeResultes{
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
