package cmdutil

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// StartSpinner displays a message with an animated spinner that updates in-place.
// The spinner continues until either the context is cancelled or the returned
// stop function is called.
func StartSpinner(ctx context.Context, w io.Writer, msg string) func() {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	frames := []rune{'|', '/', '-', '\\'}
	// animation tick; small interval for smooth spinner
	ticker := time.NewTicker(120 * time.Millisecond)

	go func() {
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-ctx.Done():
				clear := "\r" + strings.Repeat(" ", len(msg)+4) + "\r\n"
				fmt.Fprint(w, clear)
				close(done)
				return
			case <-ticker.C:
				frame := frames[i%len(frames)]
				fmt.Fprintf(w, "\r%s %c", msg, frame)
				i++
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}
