package cmdutil

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// StartSpinner displays a message with an animated spinner that updates in-place.
// The spinner continues until either the context is cancelled or the returned
// stop function is called.
func StartSpinner(ctx context.Context, w io.Writer, msg string) func() {

	// msg should not have any new lines because this will break the spinner display.
	msg = strings.ReplaceAll(msg, "\n", " ")

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	var once sync.Once

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

	// Sync catch allows for idempotent stopping
	stop := func() {
		once.Do(func() {
			cancel()
			<-done
		})
	}

	return stop
}

// WithSpinner is a execution env wrapper that starts a spinner, runs the provided action, and ensures the
// spinner is stopped once the action completes or panics. It returns the action's returned error (if any).
func WithSpinner(ctx context.Context, w io.Writer, msg string, action func(context.Context) error) (err error) {
	stop := StartSpinner(ctx, w, msg)
	defer stop()

	return action(ctx)
}
