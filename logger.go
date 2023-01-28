package jet

import "context"

// LogFunc can be set on the Db instance to allow query logging.
type LogFunc func(ctx context.Context, queryId, query string, args ...interface{})

func NoopLogFunc(_ context.Context, _, _ string, _ ...interface{}) {
}
