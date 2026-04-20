package webcontent

import _ "embed"

// IndexHTML is the project homepage (Chinese + Agent skill setup), embedded into the binary.
//
//go:embed index.html
var IndexHTML []byte

// AdminHTML is the admin console (config + index view), embedded into the binary.
//
//go:embed admin.html
var AdminHTML []byte

// AdminOpenSearchHTML is the admin OpenSearch browse + try-search page.
//
//go:embed admin_opensearch.html
var AdminOpenSearchHTML []byte
