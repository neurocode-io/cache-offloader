package probes

import "context"

type readinessChecker struct{}

func NewReadinessChecker() readinessChecker {
	return readinessChecker{}
}

func (r readinessChecker) CheckConnection(ctx context.Context) error {
	return nil
}
