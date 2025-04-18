package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/vinothnada/web-analyzer/internal/types"
)

// URL validation regex
var urlRegex = regexp.MustCompile(`^https?://`)

// GetResults handles the incoming HTTP request to analyze a URL
func GetResults(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Setting response headers")
	setResponseHeaders(w)

	// Handle OPTIONS request for CORS
	if r.Method == http.MethodOptions {
		logrus.Info("Handling OPTIONS request")
		handleOptionsRequest(w)
		return
	}

	logrus.Info("Received analysis request")

	// Only accept POST requests
	if r.Method != http.MethodPost {
		logrus.Warn("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request payload
	payload, err := parsePayload(r)
	if err != nil {
		logrus.Error("Failed to parse payload: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the URL format
	if !isValidURL(payload.URL) {
		logrus.Warn("Invalid URL format: ", payload.URL)
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	logrus.Info("Starting page analysis for URL: ", payload.URL)

	// Analyze the page
	result, err := analyzePage(payload.URL)
	if err != nil {
		logrus.Error("Error analyzing page: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.Info("Successfully analyzed page")

	// Return the analysis result as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// setResponseHeaders sets the necessary CORS headers for the response
func setResponseHeaders(w http.ResponseWriter) {
	logrus.Debug("Setting CORS headers")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS") // Allow POST and OPTIONS
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// handleOptionsRequest handles the OPTIONS request method
func handleOptionsRequest(w http.ResponseWriter) {
	logrus.Debug("OPTIONS request received, responding with status OK")
	w.WriteHeader(http.StatusOK)
}

// parsePayload parses the incoming JSON payload from the request body
func parsePayload(r *http.Request) (types.RequestPayload, error) {
	logrus.Debug("Parsing request payload")
	var payload types.RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return payload, fmt.Errorf("Invalid request payload")
	}
	logrus.Debug("Payload parsed successfully")
	return payload, nil
}

// isValidURL checks if the provided URL is valid using regex
func isValidURL(targetURL string) bool {
	logrus.Debug("Validating URL: ", targetURL)
	valid := urlRegex.MatchString(targetURL)

	if valid {
		logrus.Debug("URL is valid")
	} else {
		logrus.Warn("URL is invalid")
	}
	return valid
}

// analyzePage analyzes the content of the page at the given URL
func analyzePage(targetURL string) (*types.AnalyzeResultes, error) {
	logrus.Info("Fetching URL: ", targetURL)
	resp, err := fetchURL(targetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := parseHTML(resp.Body)
	if err != nil {
		return nil, err
	}

	logrus.Info("Extracting data from page")
	result := &types.AnalyzeResultes{Headings: make(map[string]int)}

	var waitGroup sync.WaitGroup
	internalLinks, externalLinks := 0, 0
	accessibleExternalLinks, brokenExternalLinks := 0, 0
	linksChan := make(chan string)
	statusChan := make(chan string)

	result.HTMLVersion = getHtmlVersion(doc)
	result.Title = extractTitle(doc)
	result.Headings = countHeadings(doc)

	parsedURL, _ := url.Parse(targetURL)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		hrefParsed, err := url.Parse(href)
		if err != nil {
			return
		}

		if hrefParsed.Host == "" || hrefParsed.Host == parsedURL.Host {
			internalLinks++
		} else {
			externalLinks++
			waitGroup.Add(1)
			go func(link string) {
				defer waitGroup.Done()
				linksChan <- link
			}(href)
		}
	})

	go func() {
		waitGroup.Wait()
		close(linksChan)
	}()

	go func() {
		for link := range linksChan {
			statusChan <- checkLinkAccessibility(link)
		}
		close(statusChan)
	}()

	for status := range statusChan {
		if status == "accessible" {
			accessibleExternalLinks++
		} else {
			brokenExternalLinks++
		}
	}

	result.InternalLinks = internalLinks
	result.ExternalLinks = externalLinks
	result.AccessibleExternalLinks = accessibleExternalLinks
	result.BrokenExternalLinks = brokenExternalLinks
	result.HasLoginForm = hasLoginForm(doc)

	logrus.Info("Page analysis completed successfully")
	return result, nil
}

func getHtmlVersion(doc *goquery.Document) string {
	rawHTML, err := doc.Html()
	if err != nil {
		logrus.Error("Error extracting HTML content:", err)
		return ""
	}
	rawHTML = strings.TrimSpace(rawHTML)
	if strings.Contains(rawHTML, "<!DOCTYPE html>") {
		return "HTML5"
	}
	if strings.Contains(rawHTML, "<!DOCTYPE HTML PUBLIC") {
		return "HTML 4"
	}
	if strings.Contains(rawHTML, "<!DOCTYPE html PUBLIC") {
		return "XHTML"
	}
	return "Unknown HTML version"
}

// fetchURL sends a GET request to fetch the URL's content
func fetchURL(targetURL string) (*http.Response, error) {
	logrus.Debug("Sending GET request to URL: ", targetURL)
	resp, err := http.Get(targetURL)
	if err != nil {
		logrus.Error("Failed to fetch URL: ", err)
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Check if the status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		logrus.Warn("URL returned non-OK status: ", resp.StatusCode)
		return nil, fmt.Errorf("URL returned status code %d", resp.StatusCode)
	}

	logrus.Debug("URL fetched successfully with status code: ", resp.StatusCode)
	return resp, nil
}

// parseHTML parses the HTML content from the response body
func parseHTML(body interface{}) (*goquery.Document, error) {
	logrus.Debug("Parsing HTML document")
	reader, ok := body.(io.Reader)
	if !ok {
		logrus.Error("Failed to cast body to io.Reader")
		return nil, fmt.Errorf("failed to cast body to io.Reader")
	}

	// Parse the HTML using goquery
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		logrus.Error("Failed to parse HTML: ", err)
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	logrus.Debug("HTML document parsed successfully")
	return doc, nil
}

// extractTitle extracts the title of the page from the HTML document
func extractTitle(doc *goquery.Document) string {
	logrus.Debug("Extracting title from HTML")
	return doc.Find("title").Text()
}

// countHeadings counts the number of each heading (h1-h6) tags on the page
func countHeadings(doc *goquery.Document) map[string]int {
	logrus.Debug("Counting headings (h1 to h6)")
	headings := make(map[string]int)
	for i := 1; i <= 6; i++ {
		hTag := fmt.Sprintf("h%d", i)
		headings[hTag] = doc.Find(hTag).Length()
	}
	logrus.Debug("Headings counted successfully")
	return headings
}

func checkLinkAccessibility(link string) string {
	resp, err := http.Head(link)
	if err != nil {
		logrus.Debug("Error checking link:", link, " Error:", err)
		return "broken"
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "accessible"
	}
	return "broken"
}

// hasLoginForm checks if the page contains a login form
func hasLoginForm(doc *goquery.Document) bool {
	logrus.Debug("Checking for login form")

	// Check for password field
	if doc.Find("input[type='password']").Length() > 0 {
		return true
	}

	// Check for common social media login buttons (Facebook, Google, etc.)
	if doc.Find("button, a").FilterFunction(func(i int, s *goquery.Selection) bool {
		text := s.Text()
		return strings.Contains(strings.ToLower(text), "login with") || strings.Contains(strings.ToLower(text), "sign in with")
	}).Length() > 0 {
		return true
	}
	return false

}
