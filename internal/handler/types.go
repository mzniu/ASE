package handler

// searchRequest is the JSON body for POST /v1/search.
type searchRequest struct {
	Query     string   `json:"query"`
	Providers []string `json:"providers"` // optional: e.g. ["baidu"], ["bing"], ["baidu","bing","tavily"]; omit to use server default
	// DeepSearch 省略时沿用环境变量 PROVIDER_FETCH_RESULT_URLS；若传入则按本次请求是否对 Provider 结果 URL 再抓取正文。
	DeepSearch *bool `json:"deepsearch,omitempty"`
}

// documentIndexRequest is the JSON body for POST /v1/documents.
type documentIndexRequest struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	BodyText string `json:"body_text"`
}
