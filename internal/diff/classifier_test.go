package diff

import "testing"

func TestClassify_CRDRemoved(t *testing.T) {
	c := Classify(Change{Path: "crds.mycrd.example.com", Kind: Removed, OldValue: "mycrd.example.com"})
	if c.Risk != RiskCritical {
		t.Errorf("CRD removal should be CRITICAL, got %s", c.Risk)
	}
}

func TestClassify_CRDAdded(t *testing.T) {
	c := Classify(Change{Path: "crds.newcrd.example.com", Kind: Added, NewValue: "newcrd.example.com"})
	if c.Risk != RiskHigh {
		t.Errorf("CRD addition should be HIGH, got %s", c.Risk)
	}
}

func TestClassify_DependencyAdded(t *testing.T) {
	c := Classify(Change{Path: "dependencies.cert-manager", Kind: Added})
	if c.Risk != RiskHigh {
		t.Errorf("dependency addition should be HIGH, got %s", c.Risk)
	}
}

func TestClassify_DependencyRemoved(t *testing.T) {
	c := Classify(Change{Path: "dependencies.old-dep", Kind: Removed})
	if c.Risk != RiskHigh {
		t.Errorf("dependency removal should be HIGH, got %s", c.Risk)
	}
}

func TestClassify_ResourceLimits(t *testing.T) {
	c := Classify(Change{Path: "controller.resources.limits.cpu", Kind: Changed, OldValue: "100m", NewValue: "500m"})
	if c.Risk != RiskHigh {
		t.Errorf("resource limits change should be HIGH, got %s", c.Risk)
	}
}

func TestClassify_ResourceRequests(t *testing.T) {
	c := Classify(Change{Path: "spec.containers[0].resources.requests.memory", Kind: Changed})
	if c.Risk != RiskHigh {
		t.Errorf("resource requests change should be HIGH, got %s", c.Risk)
	}
}

func TestClassify_SecurityContext(t *testing.T) {
	c := Classify(Change{Path: "spec.securityContext.runAsNonRoot", Kind: Changed})
	if c.Risk != RiskHigh {
		t.Errorf("securityContext change should be HIGH, got %s", c.Risk)
	}
}

func TestClassify_RBAC(t *testing.T) {
	c := Classify(Change{Path: "rbac.create", Kind: Changed, OldValue: true, NewValue: false})
	if c.Risk != RiskHigh {
		t.Errorf("RBAC change should be HIGH, got %s", c.Risk)
	}
}

func TestClassify_ServiceAccount(t *testing.T) {
	c := Classify(Change{Path: "serviceAccount.name", Kind: Changed})
	if c.Risk != RiskHigh {
		t.Errorf("serviceAccount change should be HIGH, got %s", c.Risk)
	}
}

func TestClassify_ImageTag(t *testing.T) {
	c := Classify(Change{Path: "controller.image.tag", Kind: Changed, OldValue: "v1.0", NewValue: "v2.0"})
	if c.Risk != RiskMedium {
		t.Errorf("image.tag change should be MEDIUM, got %s", c.Risk)
	}
}

func TestClassify_ImageRepository(t *testing.T) {
	c := Classify(Change{Path: "image.repository", Kind: Changed})
	if c.Risk != RiskMedium {
		t.Errorf("image.repository change should be MEDIUM, got %s", c.Risk)
	}
}

func TestClassify_IngressEnabled(t *testing.T) {
	c := Classify(Change{Path: "ingress.enabled", Kind: Changed, OldValue: false, NewValue: true})
	if c.Risk != RiskMedium {
		t.Errorf("ingress.enabled change should be MEDIUM, got %s", c.Risk)
	}
}

func TestClassify_ServiceType(t *testing.T) {
	c := Classify(Change{Path: "service.type", Kind: Changed, OldValue: "ClusterIP", NewValue: "LoadBalancer"})
	if c.Risk != RiskMedium {
		t.Errorf("service.type change should be MEDIUM, got %s", c.Risk)
	}
}

func TestClassify_AppVersion(t *testing.T) {
	c := Classify(Change{Path: "appVersion", Kind: Changed, OldValue: "1.0", NewValue: "2.0"})
	if c.Risk != RiskMedium {
		t.Errorf("appVersion change should be MEDIUM, got %s", c.Risk)
	}
}

func TestClassify_Annotations(t *testing.T) {
	c := Classify(Change{Path: "podAnnotations.prometheus.io/scrape", Kind: Added})
	if c.Risk != RiskLow {
		t.Errorf("annotation change should be LOW, got %s", c.Risk)
	}
}

func TestClassify_RawDiff(t *testing.T) {
	c := Classify(Change{Path: "(raw diff)", Kind: Changed})
	if c.Risk != RiskMedium {
		t.Errorf("raw diff should be MEDIUM, got %s", c.Risk)
	}
}

func TestClassify_DefaultLow(t *testing.T) {
	c := Classify(Change{Path: "someRandomKey", Kind: Changed, OldValue: "a", NewValue: "b"})
	if c.Risk != RiskLow {
		t.Errorf("unknown path change should default to LOW, got %s", c.Risk)
	}
}

func TestClassify_DescriptionSet(t *testing.T) {
	c := Classify(Change{Path: "controller.resources.limits.cpu", Kind: Changed, OldValue: "100m", NewValue: "500m"})
	if c.Description == "" {
		t.Error("expected Description to be set after Classify")
	}
}

func TestClassify_PreservesExistingDescription(t *testing.T) {
	c := Classify(Change{Path: "crds.foo", Kind: Removed, Description: "custom description"})
	if c.Description != "custom description" {
		t.Errorf("Classify should not overwrite existing description, got %q", c.Description)
	}
}

func TestClassifyAll(t *testing.T) {
	changes := []Change{
		{Path: "crds.foo", Kind: Removed},
		{Path: "image.tag", Kind: Changed},
		{Path: "someKey", Kind: Added},
	}
	result := ClassifyAll(changes)
	if result[0].Risk != RiskCritical {
		t.Errorf("expected CRITICAL, got %s", result[0].Risk)
	}
	if result[1].Risk != RiskMedium {
		t.Errorf("expected MEDIUM, got %s", result[1].Risk)
	}
	if result[2].Risk != RiskLow {
		t.Errorf("expected LOW, got %s", result[2].Risk)
	}
}

func TestRiskLevelOrdering(t *testing.T) {
	if RiskLow >= RiskMedium || RiskMedium >= RiskHigh || RiskHigh >= RiskCritical {
		t.Error("risk levels are not ordered correctly")
	}
}
