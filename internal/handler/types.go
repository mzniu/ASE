package handler

// searchRequest is the JSON body for POST /v1/search.
type searchRequest struct {
	Query     string   `json:"query"`
	Providers []string `json:"providers"` // optional: e.g. ["baidu"], ["bing"], ["baidu","bing","tavily"]; omit to use server default
}

// documentIndexRequest is the JSON body for POST /v1/documents.
type documentIndexRequest struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	BodyText string `json:"body_text"`
}
