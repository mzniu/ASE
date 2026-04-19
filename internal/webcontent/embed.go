package webcontent

import _ "embed"

// IndexHTML is the project homepage (Chinese + Agent skill setup), embedded into the binary.
//
//go:embed index.html
var IndexHTML []byte
