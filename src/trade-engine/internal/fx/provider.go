package fx

import "context"

type Provider interface {
	Rate(context.Context, string, string) (float64, error)
	Close() error
}
