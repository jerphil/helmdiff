package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func Load(dir string) (*Chart, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading chart dir: %w", err)
	}

	// The unpacked chart may be in a subdirectory (helm pull --untar creates chart-name/ inside untardir)
	chartDir := dir
	if len(entries) == 1 && entries[0].IsDir() {
		chartDir = filepath.Join(dir, entries[0].Name())
	}

	c := &Chart{Dir: chartDir}

	if err := c.loadMeta(); err != nil {
		return nil, err
	}
	if err := c.loadValues(); err != nil {
		return nil, err
	}
	if err := c.loadTemplates(); err != nil {
		return nil, err
	}
	if err := c.loadCRDs(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Chart) loadMeta() error {
	data, err := os.ReadFile(filepath.Join(c.Dir, "Chart.yaml"))
	if err != nil {
		return fmt.Errorf("reading Chart.yaml: %w", err)
	}
	return yaml.Unmarshal(data, &c.Meta)
}

func (c *Chart) loadValues() error {
	data, err := os.ReadFile(filepath.Join(c.Dir, "values.yaml"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading values.yaml: %w", err)
	}
	c.RawValues = data
	return yaml.Unmarshal(data, &c.Values)
}

func (c *Chart) loadTemplates() error {
	tplDir := filepath.Join(c.Dir, "templates")
	if _, err := os.Stat(tplDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(tplDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name := strings.TrimPrefix(path, tplDir+string(filepath.Separator))
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".tpl") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", name, err)
		}
		c.Templates = append(c.Templates, Template{Name: name, Content: content})
		return nil
	})
}

func (c *Chart) loadCRDs() error {
	crdDir := filepath.Join(c.Dir, "crds")
	if _, err := os.Stat(crdDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(crdDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		crd, err := parseCRD(path, data)
		if err != nil {
			return nil // best-effort
		}
		c.CRDs = append(c.CRDs, crd)
		return nil
	})
}

func parseCRD(path string, data []byte) (CRD, error) {
	var obj struct {
		Metadata struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
		Spec struct {
			Group string `yaml:"group"`
			Names struct {
				Kind string `yaml:"kind"`
			} `yaml:"names"`
			Versions []struct {
				Name string `yaml:"name"`
			} `yaml:"versions"`
		} `yaml:"spec"`
	}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return CRD{}, err
	}
	versions := make([]string, 0, len(obj.Spec.Versions))
	for _, v := range obj.Spec.Versions {
		versions = append(versions, v.Name)
	}
	return CRD{
		Name:     obj.Metadata.Name,
		Filename: filepath.Base(path),
		Group:    obj.Spec.Group,
		Kind:     obj.Spec.Names.Kind,
		Versions: versions,
	}, nil
}
