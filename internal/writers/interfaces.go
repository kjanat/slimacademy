package writers

import (
	"io"

	"github.com/kjanat/slimacademy/internal/streaming"
)

// Writer handles a stream of events and generates format-specific output (legacy interface)
// Deprecated: Use WriterV2 from registry.go instead
type Writer interface {
	// Handle processes a single event
	Handle(event streaming.Event)

	// Result returns the final output string
	Result() string

	// Reset clears the writer state for reuse
	Reset()
}

// StreamWriter is a Writer that outputs to an io.Writer (legacy interface)
// Deprecated: Use WriterV2 from registry.go instead
type StreamWriter interface {
	Writer

	// SetOutput sets the output destination
	SetOutput(io.Writer)
}
