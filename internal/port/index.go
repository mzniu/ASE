package port

import "context"

// IndexRepository reads from the self-hosted OpenSearch index and may index documents when backed by OpenSearch.
type IndexRepository interface {
	Search(ctx context.Context, query string) ([]Hit, error)
	// IndexDocument creates or replaces a document by id. noopindex returns ErrIndexingDisabled.
	IndexDocument(ctx context.Context, id, title, body string) error
}
