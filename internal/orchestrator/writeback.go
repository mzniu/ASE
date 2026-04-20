package orchestrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/example/ase/internal/domain"
	"github.com/example/ase/internal/port"
)

func queryWriteBackDocID(prefix, query string) string {
	p := strings.TrimSpace(prefix)
	if p == "" {
		p = "ase-q-"
	}
	q := strings.TrimSpace(query)
	sum := sha256.Sum256([]byte(q))
	return p + hex.EncodeToString(sum[:])
}

func (s *Service) acquireWriteBackSlot() {
	s.writeBackSemOnce.Do(func() {
		n := s.Config.SearchIndexWriteBackMaxConcurrency
		if n <= 0 {
			n = 16
		}
		s.writeBackSem = make(chan struct{}, n)
	})
	s.writeBackSem <- struct{}{}
}

func (s *Service) releaseWriteBackSlot() {
	<-s.writeBackSem
}

// scheduleProviderIndexWriteBack enqueues a best-effort async OpenSearch upsert (same query → same document id).
func (s *Service) scheduleProviderIndexWriteBack(query, markdown, rid string, allowWB bool) {
	if !allowWB {
		if !s.Config.SearchIndexWriteBackEnabled {
			indexWriteBackTotal.WithLabelValues("skipped_global_off").Inc()
		} else {
			indexWriteBackTotal.WithLabelValues("skipped_request_optout").Inc()
		}
		return
	}
	q := strings.TrimSpace(query)
	if q == "" {
		indexWriteBackTotal.WithLabelValues("skipped_empty_query").Inc()
		return
	}
	prefix := strings.TrimSpace(s.Config.SearchIndexWriteBackIDPrefix)
	if prefix == "" {
		prefix = "ase-q-"
	}
	body := domain.WritebackBodyFromMarkdown(markdown, s.Config.SearchIndexWriteBackMaxBodyRunes)
	if utf8.RuneCountInString(strings.TrimSpace(body)) < s.Config.SearchIndexWriteBackMinBodyRunes {
		indexWriteBackTotal.WithLabelValues("skipped_body_short").Inc()
		return
	}
	titleMax := s.Config.SearchIndexWriteBackTitleMaxRunes
	if titleMax <= 0 {
		titleMax = 256
	}
	title := domain.WritebackTitleFromQuery(q, titleMax)
	id := queryWriteBackDocID(prefix, q)
	idx := s.Index
	timeout := s.Config.SearchIndexWriteBackTimeout

	go func() {
		s.acquireWriteBackSlot()
		defer s.releaseWriteBackSlot()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := idx.IndexDocument(ctx, id, title, body)
		if err != nil {
			if errors.Is(err, port.ErrIndexingDisabled) {
				indexWriteBackTotal.WithLabelValues("noop").Inc()
				slog.Debug("index write-back skipped", "request_id", rid, "reason", "indexing_disabled")
				return
			}
			indexWriteBackTotal.WithLabelValues("error").Inc()
			slog.Warn("index write-back failed", "request_id", rid, "id", id, "err", err)
			return
		}
		indexWriteBackTotal.WithLabelValues("ok").Inc()
		slog.Info("index write-back ok", "request_id", rid, "id", id)
	}()
}
