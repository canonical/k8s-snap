package dqlite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/go-dqlite/app"
)

// nextDqlitePort is used in withDqliteCluster() to pick unique port numbers for the dqlite nodes.
var nextDqlitePort = 37312

// withDqliteCluster creates a temporary dqlite cluster of the desired size, and is meant to be
// used in tests for *dqlite.Client.
//
// Example usage:
//
// ```
//
//	func TestDqliteSomething(t *testing.T) {
//		withDqliteCluster(t, 3, func(ctx context.Context, dirs []string) {
//			fmt.Println("I have 3 nodes, directories are in %v", dirs)
//
//				// ...
//			})
//		}
//
// ```.
func withDqliteCluster(t *testing.T, size int, f func(ctx context.Context, dirs []string)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if size < 1 {
		panic(fmt.Sprintf("dqlite cluster size %v must be at least 1", size))
	}

	var dirs []string
	firstPort := nextDqlitePort
	for i := 0; i < size; i++ {
		dir := t.TempDir()
		options := []app.Option{app.WithAddress(fmt.Sprintf("127.0.0.1:%d", nextDqlitePort))}
		nextDqlitePort++
		if i > 0 {
			options = append(options, app.WithCluster([]string{fmt.Sprintf("127.0.0.1:%d", firstPort)}))
		}
		node, err := app.New(dir, options...)
		if err != nil {
			t.Fatalf("Failed to create dqlite node %d: %v", i, err)
		}
		if err := node.Ready(ctx); err != nil {
			t.Fatalf("Failed to start dqlite node %d: %v", i, err)
		}

		dirs = append(dirs, dir)
	}

	f(ctx, dirs)
}
