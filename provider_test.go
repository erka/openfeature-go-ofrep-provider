package ofrep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-feature/go-sdk/openfeature"

	"github.com/open-feature/go-sdk-contrib/providers/ofrep/internal/outbound"
)

func TestConfigurations(t *testing.T) {
	t.Run("validate header provider", func(t *testing.T) {
		c := outbound.Configuration{}

		WithHeaderProvider(func() (key string, value string) {
			return "HEADER", "VALUE"
		})(&c)

		h, v := c.Callbacks[0]()

		if h != "HEADER" {
			t.Errorf("expected header %s, but got %s", "HEADER", h)
		}

		if v != "VALUE" {
			t.Errorf("expected value %s, but got %s", "VALUE", v)
		}
	})

	t.Run("validate bearer token", func(t *testing.T) {
		c := outbound.Configuration{}

		WithBearerToken("TOKEN")(&c)

		h, v := c.Callbacks[0]()

		if h != "Authorization" {
			t.Errorf("expected header %s, but got %s", "Authorization", h)
		}

		if v != "Bearer TOKEN" {
			t.Errorf("expected value %s, but got %s", "Bearer TOKEN", v)
		}
	})

	t.Run("validate api auth key", func(t *testing.T) {
		c := outbound.Configuration{}

		WithApiKeyAuth("TOKEN")(&c)

		h, v := c.Callbacks[0]()

		if h != "X-API-Key" {
			t.Errorf("expected header %s, but got %s", "X-API-Key", h)
		}

		if v != "TOKEN" {
			t.Errorf("expected value %s, but got %s", "TOKEN", v)
		}
	})
}

func TestWiringE2E(t *testing.T) {
	server := newMockServer(t, `{"value":true,"key":"my-flag","reason":"STATIC","variant":"true","metadata":{}}`)

	// custom client with reduced timeout
	customClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	provider := NewProvider(server.URL, WithClient(customClient))
	booleanEvaluation := provider.BooleanEvaluation(context.Background(), "flag", false, nil)

	if booleanEvaluation.Value != true {
		t.Errorf("expected %v, but got %v", true, booleanEvaluation.Value)
	}

	if booleanEvaluation.Variant != "true" {
		t.Errorf("expected %v, but got %v", "true", booleanEvaluation.Variant)
	}

	if booleanEvaluation.Reason != openfeature.StaticReason {
		t.Errorf("expected %v, but got %v", openfeature.StaticReason, booleanEvaluation.Reason)
	}

	if booleanEvaluation.Error() != nil {
		t.Errorf("expected no errors, but got %v", booleanEvaluation.Error())
	}
}

type mockHandler struct {
	response string
	t        *testing.T
}

func (r mockHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
	_, err := resp.Write([]byte(r.response))
	if err != nil {
		r.t.Logf("error writing bytes: %v", err)
	}
}

func newMockServer(t *testing.T, response string) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(mockHandler{response: response, t: t})
	t.Cleanup(s.Close)
	return s
}

func TestStringEvaluation(t *testing.T) {
	server := newMockServer(t, `{"value":"hello","key":"my-flag","reason":"STATIC","variant":"variant","metadata":{}}`)
	provider := NewProvider(server.URL, WithClient(&http.Client{Timeout: 1 * time.Second}))
	result := provider.StringEvaluation(context.Background(), "flag", "default", nil)

	if result.Value != "hello" {
		t.Errorf("expected %q, got %q", "hello", result.Value)
	}
	if result.Variant != "variant" {
		t.Errorf("expected %q, got %q", "variant", result.Variant)
	}
	if result.Reason != openfeature.StaticReason {
		t.Errorf("expected %v, got %v", openfeature.StaticReason, result.Reason)
	}
	if result.Error() != nil {
		t.Errorf("expected no error, got %v", result.Error())
	}
}

func TestFloatEvaluation(t *testing.T) {
	server := newMockServer(t, `{"value":3.14,"key":"my-flag","reason":"STATIC","variant":"variant","metadata":{}}`)
	provider := NewProvider(server.URL, WithClient(&http.Client{Timeout: 1 * time.Second}))
	result := provider.FloatEvaluation(context.Background(), "flag", 0.0, nil)

	if result.Value != 3.14 {
		t.Errorf("expected %v, got %v", 3.14, result.Value)
	}
	if result.Variant != "variant" {
		t.Errorf("expected %q, got %q", "variant", result.Variant)
	}
	if result.Reason != openfeature.StaticReason {
		t.Errorf("expected %v, got %v", openfeature.StaticReason, result.Reason)
	}
	if result.Error() != nil {
		t.Errorf("expected no error, got %v", result.Error())
	}
}

func TestIntEvaluation(t *testing.T) {
	server := newMockServer(t, `{"value":42,"key":"my-flag","reason":"STATIC","variant":"variant","metadata":{}}`)
	provider := NewProvider(server.URL, WithClient(&http.Client{Timeout: 1 * time.Second}))
	result := provider.IntEvaluation(context.Background(), "flag", 0, nil)

	if result.Value != 42 {
		t.Errorf("expected %v, got %v", 42, result.Value)
	}
	if result.Variant != "variant" {
		t.Errorf("expected %q, got %q", "variant", result.Variant)
	}
	if result.Reason != openfeature.StaticReason {
		t.Errorf("expected %v, got %v", openfeature.StaticReason, result.Reason)
	}
	if result.Error() != nil {
		t.Errorf("expected no error, got %v", result.Error())
	}
}

func TestObjectEvaluation(t *testing.T) {
	server := newMockServer(t, `{"value":{"key":"val"},"key":"my-flag","reason":"STATIC","variant":"variant","metadata":{}}`)
	provider := NewProvider(server.URL, WithClient(&http.Client{Timeout: 1 * time.Second}))
	result := provider.ObjectEvaluation(context.Background(), "flag", nil, nil)

	if result.Error() != nil {
		t.Errorf("expected no error, got %v", result.Error())
	}
	if result.Variant != "variant" {
		t.Errorf("expected %q, got %q", "variant", result.Variant)
	}
	if result.Reason != openfeature.StaticReason {
		t.Errorf("expected %v, got %v", openfeature.StaticReason, result.Reason)
	}

	m, ok := result.Value.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result.Value)
	}
	if m["key"] != "val" {
		t.Errorf("expected %q, got %v", "val", m["key"])
	}
}

func TestWithFromEnv(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		initialConfig outbound.Configuration
		baseURI       string
		wantBaseURI   string
		wantTimeout   time.Duration
		wantHeaders   map[string]string
	}{
		{
			name: "configure endpoint from env",
			envVars: map[string]string{
				"OFREP_ENDPOINT": "http://test.example.com",
			},
			initialConfig: outbound.Configuration{},
			wantBaseURI:   "http://test.example.com",
		},
		{
			name: "configure timeout from env with raw milliseconds",
			envVars: map[string]string{
				"OFREP_TIMEOUT_MS": "3000",
			},
			initialConfig: outbound.Configuration{},
			wantTimeout:   3 * time.Second,
		},
		{
			name: "configure timeout from env with negative milliseconds",
			envVars: map[string]string{
				"OFREP_TIMEOUT_MS": "-5000",
			},
			initialConfig: outbound.Configuration{Timeout: 33 * time.Second},
			wantTimeout:   33 * time.Second,
		},
		{
			name: "ignore invalid timeout",
			envVars: map[string]string{
				"OFREP_TIMEOUT_MS": "invalid",
			},
			initialConfig: outbound.Configuration{Timeout: 10 * time.Second},
			wantTimeout:   10 * time.Second,
		},
		{
			name: "configure custom headers from env",
			envVars: map[string]string{
				"OFREP_HEADERS": "X-Custom-1=Value1,X-Custom-2=Value2",
			},
			initialConfig: outbound.Configuration{},
			wantHeaders: map[string]string{
				"X-Custom-1": "Value1",
				"X-Custom-2": "Value2",
			},
		},
		{
			name: "configure all options from env",
			envVars: map[string]string{
				"OFREP_ENDPOINT":   "http://all.example.com",
				"OFREP_TIMEOUT_MS": "3000",
				"OFREP_HEADERS":    "X-Test=TestValue",
			},
			initialConfig: outbound.Configuration{},
			wantBaseURI:   "http://all.example.com",
			wantTimeout:   3 * time.Second,
			wantHeaders: map[string]string{
				"X-Test": "TestValue",
			},
		},
		{
			name: "empty env variables do not override defaults",
			envVars: map[string]string{
				"OFREP_ENDPOINT":   "",
				"OFREP_TIMEOUT_MS": "",
			},
			initialConfig: outbound.Configuration{
				BaseURI: "http://default.example.com",
				Timeout: 15 * time.Second,
			},
			wantBaseURI: "http://default.example.com",
			wantTimeout: 15 * time.Second,
		},
		{
			name: "configure headers with baggage header format",
			envVars: map[string]string{
				"OFREP_HEADERS": "Key1=value1 ,Key2=val%3Due2,Key3=base64==,Key4 = 50%25",
			},
			initialConfig: outbound.Configuration{},
			wantHeaders: map[string]string{
				"Key1": "value1",
				"Key2": "val=ue2",
				"Key3": "base64==",
				"Key4": "50%",
			},
		},
		{
			name: "configure headers with auth header format",
			envVars: map[string]string{
				"OFREP_HEADERS": "Authorization=Bearer%20token,X-Custom=value",
			},
			initialConfig: outbound.Configuration{},
			wantHeaders: map[string]string{
				"Authorization": "Bearer token",
				"X-Custom":      "value",
			},
		},
		{
			name: "programmatic baseURI overrides env",
			envVars: map[string]string{
				"OFREP_ENDPOINT": "http://env.example.com",
			},
			initialConfig: outbound.Configuration{},
			baseURI:       "http://programmatic.example.com",
			wantBaseURI:   "http://programmatic.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			c := tt.initialConfig
			WithFromEnv()(&c)
			if tt.baseURI != "" {
				c.BaseURI = tt.baseURI
			}

			if tt.wantBaseURI != "" && c.BaseURI != tt.wantBaseURI {
				t.Errorf("expected BaseURI %s, but got %s", tt.wantBaseURI, c.BaseURI)
			}

			if tt.wantTimeout != 0 && c.Timeout != tt.wantTimeout {
				t.Errorf("expected Timeout %v, but got %v", tt.wantTimeout, c.Timeout)
			}

			actualHeaders := make(map[string]string)
			for _, cb := range c.Callbacks {
				k, v := cb()
				actualHeaders[k] = v
			}

			if tt.wantHeaders != nil {
				for expectedKey, expectedValue := range tt.wantHeaders {
					if actualValue, ok := actualHeaders[expectedKey]; !ok {
						t.Errorf("expected header %s not found", expectedKey)
					} else if actualValue != expectedValue {
						t.Errorf("expected %s=%s, but got %s=%s", expectedKey, expectedValue, expectedKey, actualValue)
					}
				}
			}

			if len(tt.wantHeaders) == 0 && len(actualHeaders) != 0 {
				t.Errorf("expected no headers, but got %v", actualHeaders)
			}
		})
	}
}
