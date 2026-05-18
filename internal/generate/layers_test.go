package generate

import (
	"reflect"
	"testing"

	"github.com/siyuqian/gocraft/internal/prompt"
)

func TestLayers(t *testing.T) {
	cases := []struct {
		name string
		cfg  prompt.Config
		want []string
	}{
		{
			name: "chi none no-sentry",
			cfg:  prompt.Config{HTTP: prompt.HTTPChi, Async: prompt.AsyncNone, Sentry: false},
			want: []string{"base", "http/chi"},
		},
		{
			name: "stdlib none sentry",
			cfg:  prompt.Config{HTTP: prompt.HTTPStdlib, Async: prompt.AsyncNone, Sentry: true},
			want: []string{"base", "http/stdlib", "obs/sentry"},
		},
		{
			name: "chi river sentry",
			cfg:  prompt.Config{HTTP: prompt.HTTPChi, Async: prompt.AsyncRiver, Sentry: true},
			want: []string{"base", "http/chi", "async/river", "obs/sentry"},
		},
		{
			name: "stdlib pool no-sentry",
			cfg:  prompt.Config{HTTP: prompt.HTTPStdlib, Async: prompt.AsyncPool, Sentry: false},
			want: []string{"base", "http/stdlib", "async/pool"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Layers(tc.cfg)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Layers() = %v, want %v", got, tc.want)
			}
		})
	}
}
