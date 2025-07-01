package writers

import (
	"io"

	"github.com/kjanat/slimacademy/internal/events"
)

// Writer handles a stream of events and generates format-specific output
type Writer interface {
	// Handle processes a single event
	Handle(event events.Event)
	
	// Result returns the final output string
	Result() string
	
	// Reset clears the writer state for reuse
	Reset()
}

// StreamWriter is a Writer that outputs to an io.Writer
type StreamWriter interface {
	Writer
	
	// SetOutput sets the output destination
	SetOutput(io.Writer)
}