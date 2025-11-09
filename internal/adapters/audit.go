package adapters

import "context"

type auditAdapter struct {
}

func NewAuditAdapter() AuditAdapter {
	return &auditAdapter{}
}

func (a *auditAdapter) Audit(ctx context.Context) error {
	//TODO implement me
	return nil
}
