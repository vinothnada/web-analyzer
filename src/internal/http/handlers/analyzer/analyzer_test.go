package analyzer

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock HTTP Client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestGetResults(t *testing.T) {
	// Create mock request and response
	tests := []struct {
		name        string
		method      string
		body        string
		wantStatus  int
		mockHandler func(*MockHTTPClient)
	}{
		{
			name:       "Valid URL",
			method:     http.MethodPost,
			body:       `{"URL": "https://example.com"}`,
			wantStatus: http.StatusOK,
			mockHandler: func(client *MockHTTPClient) {
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("<html></html>")),
				}, nil)
			},
		},
		{
			name:       "Invalid URL",
			method:     http.MethodPost,
			body:       `{"URL": "ht://example.com"}`,
			wantStatus: http.StatusBadRequest,
			mockHandler: func(client *MockHTTPClient) {
				// No HTTP call, URL validation will fail.
			},
		},
		// Add more test cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request
			req, err := http.NewRequest(tt.method, "/analyze", strings.NewReader(tt.body))
			assert.NoError(t, err)

			// Create a mock response writer
			rr := httptest.NewRecorder()

			// Setup the mock HTTP client
			mockClient := new(MockHTTPClient)
			tt.mockHandler(mockClient)

			// Call GetResults with the mock client
			GetResults(rr, req)

			// Check if the status code matches expected value
			assert.Equal(t, tt.wantStatus, rr.Code)

			// Add more checks for returned data if necessary...
		})
	}
}

func Test_handleOptionsRequest(t *testing.T) {
	// Test case for handling OPTIONS request
	tests := []struct {
		name string
	}{
		{name: "OPTIONS request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock response writer
			rr := httptest.NewRecorder()

			// Call handleOptionsRequest
			setResponseHeaders(rr)
			handleOptionsRequest(rr)

			// Assert status code and headers for CORS
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "POST, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
		})
	}
}

func Test_parsePayload(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "Valid Payload", body: `{"URL": "https://example.com"}`, wantErr: false},
		{name: "Invalid Payload", body: `{"Invalid": "field"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/analyze", strings.NewReader(tt.body))
			assert.NoError(t, err)

			got, err := parsePayload(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePayload() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && got.URL != "" {
				t.Errorf("parsePayload() = %v, want error", got)
			}
		})
	}
}

func Test_isValidURL(t *testing.T) {
	tests := []struct {
		name      string
		targetURL string
		want      bool
	}{
		{name: "Valid URL", targetURL: "https://example.com", want: true},
		{name: "Invalid URL", targetURL: "http://example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidURL(tt.targetURL)
			if got != tt.want {
				t.Errorf("isValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_analyzePage(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "Valid URL", url: "https://example.com", wantErr: false},
		{name: "Invalid URL", url: "http://example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP client
			mockClient := new(MockHTTPClient)
			mockClient.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("<html></html>")),
			}, nil)

			_, err := analyzePage(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzePage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getHtmlVersion(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{name: "HTML5", html: "<!DOCTYPE html><html><head></head><body></body></html>", want: "HTML5"},
		{name: "HTML4", html: "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01//EN\" \"http://www.w3.org/TR/html4/strict.dtd\"><html><head></head><body></body></html>", want: "HTML 4"},
		{name: "XHTML", html: "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\"><html><head></head><body></body></html>", want: "XHTML"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			got := getHtmlVersion(doc)
			if got != tt.want {
				t.Errorf("getHtmlVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "Valid URL", url: "https://example.com", wantErr: false},
		{name: "Invalid URL", url: "http://example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockHTTPClient)
			mockClient.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("<html></html>")),
			}, nil)

			_, err := fetchURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseHTML(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "Valid HTML", body: "<html><body></body></html>", wantErr: false},
		{name: "Invalid HTML", body: "<html><body></html>", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHTML(strings.NewReader(tt.body))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHTML() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && got != nil {
				t.Errorf("parseHTML() = %v, want error", got)
			}
		})
	}
}

func Test_extractTitle(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{name: "Extract Title", html: "<html><head><title>Test</title></head></html>", want: "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			got := extractTitle(doc)
			if got != tt.want {
				t.Errorf("extractTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_countHeadings(t *testing.T) {
	tests := []struct {
		name string
		html string
		want map[string]int
	}{
		{name: "Count Headings", html: "<html><h1>Heading 1</h1><h2>Heading 2</h2><h2>Heading 3</h2></html>", want: map[string]int{"h1": 1, "h2": 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			got := countHeadings(doc)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("countHeadings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkLinkAccessibility(t *testing.T) {
	tests := []struct {
		name string
		link string
		want string
	}{
		{name: "Accessible Link", link: "https://example.com", want: "accessible"},
		{name: "Broken Link", link: "https://nonexistent.com", want: "broken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkLinkAccessibility(tt.link)
			if got != tt.want {
				t.Errorf("checkLinkAccessibility() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasLoginForm(t *testing.T) {
	tests := []struct {
		name string
		html string
		want bool
	}{
		{name: "Has Login Form", html: "<html><input type='password'></html>", want: true},
		{name: "No Login Form", html: "<html><input type='text'></html>", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			got := hasLoginForm(doc)
			if got != tt.want {
				t.Errorf("hasLoginForm() = %v, want %v", got, tt.want)
			}
		})
	}
}
