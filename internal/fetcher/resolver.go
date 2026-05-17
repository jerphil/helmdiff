package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// wellKnownRepos maps common chart names to their Helm repo URLs.
var wellKnownRepos = map[string]string{
	"ingress-nginx":                "https://kubernetes.github.io/ingress-nginx",
	"cert-manager":                 "https://charts.jetstack.io",
	"external-secrets":             "https://charts.external-secrets.io",
	"prometheus":                   "https://prometheus-community.github.io/helm-charts",
	"kube-prometheus-stack":        "https://prometheus-community.github.io/helm-charts",
	"grafana":                      "https://grafana.github.io/helm-charts",
	"loki":                         "https://grafana.github.io/helm-charts",
	"tempo":                        "https://grafana.github.io/helm-charts",
	"argo-cd":                      "https://argoproj.github.io/argo-helm",
	"argocd":                       "https://argoproj.github.io/argo-helm",
	"sealed-secrets":               "https://bitnami-labs.github.io/sealed-secrets",
	"velero":                       "https://vmware-tanzu.github.io/helm-charts",
	"cluster-autoscaler":           "https://kubernetes.github.io/autoscaler",
	"metrics-server":               "https://kubernetes-sigs.github.io/metrics-server",
	"aws-load-balancer-controller": "https://aws.github.io/eks-charts",
	"eks-charts":                   "https://aws.github.io/eks-charts",
	"datadog":                      "https://helm.datadoghq.com",
	"newrelic":                     "https://helm-charts.newrelic.com",
	"vault":                        "https://helm.releases.hashicorp.com",
	"consul":                       "https://helm.releases.hashicorp.com",
	"nginx":                        "https://helm.nginx.com/stable",
	"traefik":                      "https://helm.traefik.io/traefik",
	"istio":                        "https://istio-release.storage.googleapis.com/charts",
	"linkerd":                      "https://helm.linkerd.io/stable",
	"calico":                       "https://projectcalico.docs.tigera.io/charts",
	"cilium":                       "https://helm.cilium.io",
	"longhorn":                     "https://charts.longhorn.io",
	"rook-ceph":                    "https://charts.rook.io/release",
	"mongodb":                      "https://mongodb.github.io/helm-charts",
	"mysql":                        "https://helm.mysql.com",
	"redis":                        "https://charts.bitnami.com/bitnami",
	"postgresql":                   "https://charts.bitnami.com/bitnami",
	"elasticsearch":                "https://helm.elastic.co",
	"kibana":                       "https://helm.elastic.co",
	"fluentd":                      "https://fluent.github.io/helm-charts",
	"fluent-bit":                   "https://fluent.github.io/helm-charts",
	"jaeger":                       "https://jaegertracing.github.io/helm-charts",
	"opentelemetry-collector":      "https://open-telemetry.github.io/opentelemetry-helm-charts",
	"kyverno":                      "https://kyverno.github.io/kyverno",
	"falco":                        "https://falcosecurity.github.io/charts",
	"trivy":                        "https://aquasecurity.github.io/helm-charts",
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

// resolve returns a repo URL for the given chart name, checking the well-known
// list first, then falling back to Artifact Hub.
func resolve(chart string) string {
	if url, ok := wellKnownRepos[chart]; ok {
		return url
	}
	if url, err := artifactHubLookup(chart); err == nil {
		return url
	}
	return ""
}

func artifactHubLookup(chart string) (string, error) {
	url := fmt.Sprintf("https://artifacthub.io/api/v1/packages/helm/%s/%s", chart, chart)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("artifact hub returned %d", resp.StatusCode)
	}

	var pkg struct {
		Repository struct {
			URL string `json:"url"`
		} `json:"repository"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return "", err
	}
	return pkg.Repository.URL, nil
}
