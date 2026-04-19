// Package opensearch implements port.IndexRepository against OpenSearch 2.x (DETAILED_DESIGN §6.3).
package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/example/ase/internal/adapter/noopindex"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/port"
	sdk "github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

// Repo executes multi_match queries on title + body_text.
type Repo struct {
	api   *opensearchapi.Client
	index string
	size  int
	maxQ  int
}

// NewFromConfig returns a real index client when OPENSEARCH_URLS and OPENSEARCH_INDEX are set.
// Otherwise returns noopindex.Repo{}. Client creation errors are returned for misconfiguration.
func NewFromConfig(cfg config.Config) (port.IndexRepository, error) {
	if len(cfg.OpenSearchURLs) == 0 || cfg.OpenSearchIndex == "" {
		return noopindex.Repo{}, nil
	}
	size := cfg.OpenSearchSearchSize
	if size <= 0 {
		size = 10
	}
	api, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: sdk.Config{
			Addresses: cfg.OpenSearchURLs,
			Username:    cfg.OpenSearchUser,
			Password:    cfg.OpenSearchPassword,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("opensearch client: %w", err)
	}
	return &Repo{
		api:   api,
		index: cfg.OpenSearchIndex,
		size:  size,
		maxQ:  cfg.MaxQueryRunes,
	}, nil
}

type searchRequest struct {
	Size   int      `json:"size"`
	Source []string `json:"_source"`
	Query  struct {
		MultiMatch multiMatch `json:"multi_match"`
	} `json:"query"`
}

type multiMatch struct {
	Query    string   `json:"query"`
	Fields   []string `json:"fields"`
	Type     string   `json:"type"`
	Operator string   `json:"operator,omitempty"`
}

type sourceDoc struct {
	Title    string `json:"title"`
	BodyText string `json:"body_text"`
}

// Search runs a multi_match on title^2 and body_text (REQ-F-006).
func (r *Repo) Search(ctx context.Context, query string) ([]port.Hit, error) {
	q := truncateRunes(query, r.maxQ)
	if strings.TrimSpace(q) == "" {
		return nil, nil
	}
	body := searchRequest{
		Size:   r.size,
		Source: []string{"title", "body_text"},
	}
	body.Query.MultiMatch = multiMatch{
		Query:    q,
		Fields:   []string{"title^2", "body_text"},
		Type:     "best_fields",
		Operator: "or",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := r.api.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{r.index},
		Body:    bytes.NewReader(raw),
	})
	if err != nil {
		return nil, fmt.Errorf("opensearch search: %w", err)
	}
	out := make([]port.Hit, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		var src sourceDoc
		if len(h.Source) > 0 {
			_ = json.Unmarshal(h.Source, &src)
		}
		text := composeHitBody(src.Title, src.BodyText)
		if strings.TrimSpace(text) == "" {
			continue
		}
		out = append(out, port.Hit{
			ID:    h.ID,
			Body:  text,
			Score: float64(h.Score),
		})
	}
	return out, nil
}

// IndexDocument indexes or replaces a document with fields title and body_text (same mapping as search _source).
func (r *Repo) IndexDocument(ctx context.Context, id, title, body string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("document id required")
	}
	doc := sourceDoc{Title: title, BodyText: body}
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	if _, err := r.api.Index(ctx, opensearchapi.IndexReq{
		Index:      r.index,
		DocumentID: id,
		Body:       bytes.NewReader(raw),
	}); err != nil {
		return fmt.Errorf("opensearch index: %w", err)
	}
	return nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return strings.TrimSpace(s)
	}
	r := []rune(strings.TrimSpace(s))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max])
}

func composeHitBody(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	switch {
	case title != "" && body != "":
		return "**" + title + "**\n\n" + body
	case body != "":
		return body
	default:
		return title
	}
}
