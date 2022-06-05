package probes

import "context"

type ReadinessChecker struct{}

func NewReadinessChecker() ReadinessChecker {
	return ReadinessChecker{}
}

func (r ReadinessChecker) CheckConnection(ctx context.Context) error {
	return nil
}
