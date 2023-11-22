package migrator

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8s-dqlite/kine/client"
)

func putKey(ctx context.Context, c client.Client, key string, value []byte) error {
	err := c.Create(ctx, key, value)
	if err == nil {
		return nil
	} else if err.Error() != "key exists" {
		return fmt.Errorf("failed to create key %q: %w", key, err)
	}
	// failed to create key because it exists, make a few attempts to overwrite it
	for i := 0; i < 5 && err != nil; i++ {
		time.Sleep(50 * time.Millisecond)
		err = c.Put(ctx, key, value)
	}
	if err != nil {
		return fmt.Errorf("failed to put key %q: %w", key, err)
	}
	return nil
}
