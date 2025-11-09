package adapters

import "context"

type AuditAdapter interface {
	Audit(ctx context.Context) error
}
