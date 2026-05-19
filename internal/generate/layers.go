package generate

import "github.com/siyuqian/gocraft/internal/prompt"

// Layers returns the ordered list of template layer roots for cfg.
// Order is: base, http/<choice>, async/<choice> (if not none), obs/sentry (if enabled).
// Later layers overwrite earlier layers when the same relative path appears.
func Layers(cfg prompt.Config) []string {
	out := []string{"base", "http/" + cfg.HTTP}
	if cfg.Async != prompt.AsyncNone {
		out = append(out, "async/"+cfg.Async)
	}
	if cfg.Sentry {
		out = append(out, "obs/sentry")
	}
	return out
}
