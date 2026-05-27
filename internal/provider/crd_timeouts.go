package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
)

const defaultResourceTimeout = 10 * time.Minute

func contextWithCreateTimeout(ctx context.Context, t timeouts.Value) (context.Context, context.CancelFunc) {
	duration, diags := t.Create(ctx, defaultResourceTimeout)
	if diags.HasError() {
		duration = defaultResourceTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	return ctx, cancel
}

func contextWithReadTimeout(ctx context.Context, t timeouts.Value) (context.Context, context.CancelFunc) {
	duration, diags := t.Read(ctx, defaultResourceTimeout)
	if diags.HasError() {
		duration = defaultResourceTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	return ctx, cancel
}

func contextWithUpdateTimeout(ctx context.Context, t timeouts.Value) (context.Context, context.CancelFunc) {
	duration, diags := t.Update(ctx, defaultResourceTimeout)
	if diags.HasError() {
		duration = defaultResourceTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	return ctx, cancel
}

func contextWithDeleteTimeout(ctx context.Context, t timeouts.Value) (context.Context, context.CancelFunc) {
	duration, diags := t.Delete(ctx, defaultResourceTimeout)
	if diags.HasError() {
		duration = defaultResourceTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	return ctx, cancel
}
