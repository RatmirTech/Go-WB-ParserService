package httpadapter

import (
	"context"
)

type HTTPAdapter interface {
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context)
	URL() string
}
