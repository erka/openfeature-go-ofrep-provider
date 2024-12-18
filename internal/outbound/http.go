package outbound

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	of "github.com/open-feature/go-sdk/openfeature"
)

const ofrepV1 = "/ofrep/v1/evaluate/flags"

// HeaderCallback is a callback returning header name and header value.
type HeaderCallback func() (name string, value string)

type Configuration struct {
	BaseURI               string
	Callbacks             []HeaderCallback
	Client                *http.Client
	ClientPollingInterval time.Duration
}

func (c *Configuration) PollingEnabled() bool {
	return c.ClientPollingInterval > 0
}

func (c *Configuration) PollingInterval() time.Duration {
	return c.ClientPollingInterval
}

type Resolution struct {
	Data    []byte
	Status  int
	Headers http.Header
}

// Outbound client for http communication.
type Outbound struct {
	baseURI        string
	client         *http.Client
	headerProvider []HeaderCallback
}

func NewHTTP(cfg Configuration) *Outbound {
	if cfg.Client == nil {
		cfg.Client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &Outbound{
		headerProvider: cfg.Callbacks,
		baseURI:        cfg.BaseURI,
		client:         cfg.Client,
	}
}

func (h *Outbound) Single(ctx context.Context, key string, payload []byte) (*Resolution, error) {
	path, err := url.JoinPath(h.baseURI, ofrepV1, key)
	if err != nil {
		return nil, fmt.Errorf("error building request path: %w", err)
	}
	return h.sendRequest(ctx, path, payload)
}

func (h *Outbound) Bulk(ctx context.Context, payload []byte) (*Resolution, error) {
	path, err := url.JoinPath(h.baseURI, ofrepV1)
	if err != nil {
		return nil, fmt.Errorf("error building request path: %w", err)
	}
	return h.sendRequest(ctx, path, payload)
}

func (h *Outbound) sendRequest(ctx context.Context, path string, payload []byte) (*Resolution, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(payload))
	if err != nil {
		resErr := of.NewGeneralResolutionError(fmt.Sprintf("request building error: %v", err))
		return nil, &resErr
	}

	for _, callback := range h.headerProvider {
		req.Header.Set(callback())
	}

	rsp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make a request: %w", err)
	}
	defer rsp.Body.Close()

	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the body: %w", err)
	}

	return &Resolution{
		Data:    b,
		Status:  rsp.StatusCode,
		Headers: rsp.Header,
	}, nil
}
