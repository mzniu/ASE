package port

import "errors"

// ErrIndexingDisabled is returned by noopindex when POST /v1/documents is not backed by a real index.
var ErrIndexingDisabled = errors.New("indexing disabled")
