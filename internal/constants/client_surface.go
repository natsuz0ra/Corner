package constants

import "context"

const (
	ClientSurfaceWeb = "web"
	ClientSurfaceCLI = "cli"
)

type clientSurfaceContextKey struct{}

func WithClientSurface(ctx context.Context, surface string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if surface == "" {
		return ctx
	}
	return context.WithValue(ctx, clientSurfaceContextKey{}, surface)
}

func ClientSurfaceFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	surface, _ := ctx.Value(clientSurfaceContextKey{}).(string)
	return surface
}
