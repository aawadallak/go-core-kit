package audit

import "time"

type Options func(*Orchestrator)

func WithBatchSize(size int) Options {
	return func(p *Orchestrator) {
		p.batchSize = size
	}
}

func WithBatchInterval(interval time.Duration) Options {
	return func(p *Orchestrator) {
		p.batchInterval = interval
	}
}

func WithProvider(provider Provider) Options {
	return func(o *Orchestrator) {
		o.provider = provider
	}
}
