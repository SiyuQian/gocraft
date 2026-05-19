package generate

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/siyuqian/gocraft/internal/prompt"
)

// Render walks each layer in order under fsys and writes files into outDir.
// Files ending in ".tmpl" have the suffix stripped and are rendered as
// text/template with cfg as the data context. Other files are copied
// byte-for-byte. Path segments may contain template actions (e.g. "{{.Name}}")
// which are expanded against cfg. Later layers overwrite earlier layers when
// they target the same relative output path.
func Render(cfg prompt.Config, fsys fs.FS, layers []string, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outDir, err)
	}
	for _, layer := range layers {
		if err := renderLayer(cfg, fsys, layer, outDir); err != nil {
			return fmt.Errorf("layer %q: %w", layer, err)
		}
	}
	return nil
}

func renderLayer(cfg prompt.Config, fsys fs.FS, layer, outDir string) error {
	// A layer with no files (entry missing in fsys) is silently skipped.
	if _, err := fs.Stat(fsys, layer); err != nil {
		return nil
	}
	return fs.WalkDir(fsys, layer, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(layer, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		destRel, err := expandPath(rel, cfg)
		if err != nil {
			return fmt.Errorf("path %q: %w", rel, err)
		}
		destRel = strings.TrimSuffix(destRel, ".tmpl")
		destAbs := filepath.Join(outDir, filepath.FromSlash(destRel))

		if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
			return err
		}

		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".tmpl") {
			rendered, err := execTemplate(rel, string(data), cfg)
			if err != nil {
				return fmt.Errorf("render %q: %w", rel, err)
			}
			data = rendered
		}
		return os.WriteFile(destAbs, data, 0o644)
	})
}

func expandPath(p string, cfg prompt.Config) (string, error) {
	if !strings.Contains(p, "{{") {
		return p, nil
	}
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if !strings.Contains(s, "{{") {
			continue
		}
		out, err := execTemplate("path:"+s, s, cfg)
		if err != nil {
			return "", err
		}
		segs[i] = string(out)
	}
	return path.Join(segs...), nil
}

func execTemplate(name, src string, cfg prompt.Config) ([]byte, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(src)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
