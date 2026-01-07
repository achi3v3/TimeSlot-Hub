package closer

import "context"

type Closer interface {
	Close(ctx context.Context) error
	Name() string
}

type Graceful interface {
	Shutdown(ctx context.Context) error
	Name() string
}
