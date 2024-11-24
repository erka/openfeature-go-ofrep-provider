package evaluate

import (
	"context"
	"fmt"

	"github.com/open-feature/go-sdk-contrib/providers/ofrep/internal/outbound"
	of "github.com/open-feature/go-sdk/openfeature"
)

// Flags is the flag evaluator implementation. It contains domain logic of the OpenFeature flag evaluation.
type Flags struct {
	resolver resolver
}

type resolver interface {
	resolveSingle(ctx context.Context, key string, evalCtx map[string]interface{}) (*successDto, *of.ResolutionError)
}

func NewFlagsEvaluator(cfg outbound.Configuration) *Flags {
	return &Flags{
		resolver: NewOutboundResolver(cfg),
	}
}

func (h Flags) ResolveBoolean(ctx context.Context, key string, defaultValue bool, evalCtx map[string]interface{}) of.BoolResolutionDetail {
	value, resolution := resolve(ctx, h.resolver, key, defaultValue, evalCtx, convertDefault)
	return of.BoolResolutionDetail{
		Value:                    value,
		ProviderResolutionDetail: resolution,
	}
}

func (h Flags) ResolveString(ctx context.Context, key string, defaultValue string, evalCtx map[string]interface{}) of.StringResolutionDetail {
	value, resolution := resolve(ctx, h.resolver, key, defaultValue, evalCtx, convertDefault)

	return of.StringResolutionDetail{
		Value:                    value,
		ProviderResolutionDetail: resolution,
	}
}

func (h Flags) ResolveFloat(ctx context.Context, key string, defaultValue float64, evalCtx map[string]interface{}) of.FloatResolutionDetail {
	value, resolution := resolve(ctx, h.resolver, key, defaultValue, evalCtx, convertToFloat64)
	return of.FloatResolutionDetail{
		Value:                    value,
		ProviderResolutionDetail: resolution,
	}
}

func (h Flags) ResolveInt(ctx context.Context, key string, defaultValue int64, evalCtx map[string]interface{}) of.IntResolutionDetail {
	value, resolution := resolve(ctx, h.resolver, key, defaultValue, evalCtx, convertToInt64)
	return of.IntResolutionDetail{
		Value:                    value,
		ProviderResolutionDetail: resolution,
	}
}

func (h Flags) ResolveObject(ctx context.Context, key string, defaultValue interface{}, evalCtx map[string]interface{}) of.InterfaceResolutionDetail {
	value, resolution := resolve(ctx, h.resolver, key, defaultValue, evalCtx, convertDefault)
	return of.InterfaceResolutionDetail{
		Value:                    value,
		ProviderResolutionDetail: resolution,
	}
}

type convertFunc[T any] func(v any) (T, bool)

func resolve[T any](ctx context.Context, resolver resolver, key string, defaultValue T, evalCtx map[string]interface{}, convert convertFunc[T]) (T, of.ProviderResolutionDetail) {
	evalSuccess, resolutionError := resolver.resolveSingle(ctx, key, evalCtx)
	if resolutionError != nil {
		return defaultValue, of.ProviderResolutionDetail{
			ResolutionError: *resolutionError,
			Reason:          of.ErrorReason,
		}
	}

	if evalSuccess.Reason == string(of.DisabledReason) {
		return defaultValue, of.ProviderResolutionDetail{
			Reason:       of.DisabledReason,
			Variant:      evalSuccess.Variant,
			FlagMetadata: evalSuccess.Metadata,
		}
	}

	b, ok := convert(evalSuccess.Value)
	if !ok {
		return defaultValue, of.ProviderResolutionDetail{
			ResolutionError: of.NewTypeMismatchResolutionError(fmt.Sprintf(
				"resolved value %v is not of %T type", evalSuccess.Value, defaultValue)),
			Reason: of.ErrorReason,
		}
	}

	return b, of.ProviderResolutionDetail{
		Reason:       of.Reason(evalSuccess.Reason),
		Variant:      evalSuccess.Variant,
		FlagMetadata: evalSuccess.Metadata,
	}
}

func convertDefault[T any](v any) (T, bool) {
	b, ok := v.(T)
	return b, ok
}

func convertToFloat64(v any) (float64, bool) {
	switch v := v.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func convertToInt64(v any) (int64, bool) {
	switch v := v.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		value := int64(v)
		if float64(value) != v {
			return 0, false
		}
		return value, true
	default:
		return 0, false
	}
}
