package persister

import "context"

type Element interface {
	Id() string
}

type Persister[T Element] interface {
	Save(ctx context.Context, item T) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (T, error)
	List(ctx context.Context, offset, limit int) ([]T, error)
}
