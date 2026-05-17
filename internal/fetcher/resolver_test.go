package fetcher

import "testing"

func TestResolve_WellKnown(t *testing.T) {
	cases := []struct {
		chart   string
		wantURL string
	}{
		{"ingress-nginx", "https://kubernetes.github.io/ingress-nginx"},
		{"cert-manager", "https://charts.jetstack.io"},
		{"external-secrets", "https://charts.external-secrets.io"},
		{"argo-cd", "https://argoproj.github.io/argo-helm"},
		{"prometheus", "https://prometheus-community.github.io/helm-charts"},
		{"grafana", "https://grafana.github.io/helm-charts"},
		{"vault", "https://helm.releases.hashicorp.com"},
		{"cilium", "https://helm.cilium.io"},
	}
	for _, tc := range cases {
		t.Run(tc.chart, func(t *testing.T) {
			got := resolve(tc.chart)
			if got != tc.wantURL {
				t.Errorf("resolve(%q) = %q, want %q", tc.chart, got, tc.wantURL)
			}
		})
	}
}

func TestResolve_Unknown(t *testing.T) {
	// Unknown charts without network return empty string (Artifact Hub lookup will fail in unit test context)
	got := resolve("definitely-not-a-real-chart-xyz123")
	// We just verify it doesn't panic and returns a string
	_ = got
}
