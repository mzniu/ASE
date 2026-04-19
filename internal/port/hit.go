package port

// Hit is one search hit from the index (or normalized for display).
type Hit struct {
	ID         string
	Body       string
	Score      float64 // raw OpenSearch _score
	Similarity float64 // batch min-max normalized, in [0,1]
}
