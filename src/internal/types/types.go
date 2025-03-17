package types

type AnalyzeResultes struct {
	HTMLVersion             string         `json:"htmlVersion"`
	Title                   string         `json:"title"`
	Headings                map[string]int `json:"headings"`
	InternalLinks           int            `json:"internalLinks"`
	ExternalLinks           int            `json:"externalLinks"`
	HasLoginForm            bool           `json:"hasLoginForm"`
	AccessibleExternalLinks int            `json:"accessibleExternalLinks"`
	BrokenExternalLinks     int            `json:"brokenExternalLinks"`
}

type RequestPayload struct {
	URL string `json:"url"`
}
