package types

type AnalyzeResultes struct {
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
