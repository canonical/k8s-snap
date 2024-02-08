package snap

import "context"

type snapContextKey struct{}

// SnapFromContext extracts the snap instance from the provided context.
// A panic is invoked if there is not snap instance in this context.
func SnapFromContext(ctx context.Context) Snap {
	snap, ok := ctx.Value(snapContextKey{}).(Snap)
	if !ok {
		// This should never happen as the main microcluster state context should contain the snap for k8sd.
		// Thus, panic is fine here to avoid cumbersome and unnecessary error checks on client side.
		panic("There is no snap value in the given context. Make sure that the context is wrapped with snap.ContextWithSnap.")
	}
	return snap
}

// ContextWithSnap adds a snap instance to a given context.
func ContextWithSnap(ctx context.Context, snap Snap) context.Context {
	return context.WithValue(ctx, snapContextKey{}, snap)
}
