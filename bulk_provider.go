package ofrep

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/open-feature/go-sdk-contrib/providers/ofrep/internal/evaluate"
	"github.com/open-feature/go-sdk-contrib/providers/ofrep/internal/outbound"
	of "github.com/open-feature/go-sdk/openfeature"
)

const providerName = "OFREP Bulk Provider"

func NewBulkProvider(baseURI string, options ...Option) *BulkProvider {
	cfg := outbound.Configuration{
		BaseURI:               baseURI,
		ClientPollingInterval: 30 * time.Second,
	}

	for _, option := range options {
		option(&cfg)
	}

	return &BulkProvider{
		events: make(chan of.Event, 3),
		cfg:    cfg,
		state:  of.NotReadyState,
	}
}

var (
	_ of.FeatureProvider = (*BulkProvider)(nil) // ensure BulkProvider implements FeatureProvider
	_ of.StateHandler    = (*BulkProvider)(nil) // ensure BulkProvider implements StateHandler
	_ of.EventHandler    = (*BulkProvider)(nil) // ensure BulkProvider implements EventHandler
)

type BulkProvider struct {
	Provider
	cfg        outbound.Configuration
	state      of.State
	mu         sync.RWMutex
	events     chan of.Event
	cancelFunc context.CancelFunc
}

func (p *BulkProvider) Metadata() of.Metadata {
	return of.Metadata{
		Name: providerName,
	}
}

func (p *BulkProvider) Status() of.State {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.state
}

func (p *BulkProvider) Init(evalCtx of.EvaluationContext) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != of.NotReadyState {
		// avoid reinitialization if initialized
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFunc = cancel

	client := outbound.NewHTTP(p.cfg)

	flatCtx := FlattenContext(evalCtx)

	p.state = of.ReadyState
	evaluator := evaluate.NewBulkEvaluator(client, flatCtx)
	err := evaluator.Fetch(ctx)
	if err != nil {
		err = fmt.Errorf("failed to fetch data: %w", err)
		p.state = of.ErrorState
	}

	if p.cfg.PollingEnabled() {
		p.startPolling(ctx, evaluator, p.cfg.PollingInterval())
	}

	p.evaluator = evaluator
	return err
}

func (p *BulkProvider) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
		p.cancelFunc = nil
	}
	p.state = of.NotReadyState
	p.evaluator = nil
}

func (p *BulkProvider) EventChannel() <-chan of.Event {
	return p.events
}

func (p *BulkProvider) setState(state of.State, eventType of.EventType, message string) {
	p.mu.Lock()
	p.state = state
	p.mu.Unlock()
	p.events <- of.Event{
		ProviderName: providerName, EventType: eventType,
		ProviderEventDetails: of.ProviderEventDetails{Message: message},
	}
}

func (p *BulkProvider) startPolling(ctx context.Context, evaluator *evaluate.BulkEvaluator, pollingInterval time.Duration) {
	go func() {
		ticker := time.NewTicker(pollingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := evaluator.Fetch(ctx)
				if errors.Is(err, context.Canceled) {
					// shutdown
					return
				}
				p.mu.RLock()
				oldState := p.state
				p.mu.RUnlock()

				if err != nil {
					switch oldState {
					case of.ReadyState, of.StaleState:
						p.setState(of.StaleState, of.ProviderStale, err.Error())
					case of.ErrorState:
						p.setState(of.ErrorState, of.ProviderError, err.Error())
					}
					continue
				}

				switch oldState {
				case of.ReadyState:
					p.events <- of.Event{
						ProviderName: providerName, EventType: of.ProviderConfigChange,
						ProviderEventDetails: of.ProviderEventDetails{Message: "Flags is updated"},
					}
				default:
					p.setState(of.ReadyState, of.ProviderReady, "Provider is ready")
				}

			}
		}
	}()
}

func FlattenContext(evalCtx of.EvaluationContext) of.FlattenedContext {
	flatCtx := evalCtx.Attributes()
	if evalCtx.TargetingKey() != "" {
		flatCtx[of.TargetingKey] = evalCtx.TargetingKey()
	}
	return flatCtx
}
