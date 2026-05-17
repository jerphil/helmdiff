package chart

type ChartMeta struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	AppVersion   string            `yaml:"appVersion"`
	APIVersion   string            `yaml:"apiVersion"`
	Description  string            `yaml:"description"`
	KubeVersion  string            `yaml:"kubeVersion"`
	Type         string            `yaml:"type"`
	Dependencies []Dependency      `yaml:"dependencies"`
	Annotations  map[string]string `yaml:"annotations"`
}

type Dependency struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
	Condition  string `yaml:"condition"`
	Alias      string `yaml:"alias"`
}

type CRD struct {
	Name     string
	Filename string
	Group    string
	Kind     string
	Versions []string
}

type Template struct {
	Name    string
	Content []byte
}

type Chart struct {
	Dir       string
	Meta      ChartMeta
	Values    map[string]any
	RawValues []byte
	Templates []Template
	CRDs      []CRD
}
