package manifest

import "testing"

func TestVNextDeliverablesAllowMultipleSameType(t *testing.T) {
	m := &Manifest{
		Deliverables: []Deliverable{
			{ID: "app", Type: "image", TargetID: "app-image"},
			{ID: "worker", Type: "image", TargetID: "worker-image"},
		},
	}
	if err := m.ValidateDeliverables(); err != nil {
		t.Fatal(err)
	}
	specs := m.DeliverableSpecs()
	if len(specs) != 2 || specs["app"].Type != "image" || specs["worker"].Type != "image" {
		t.Fatalf("unexpected specs: %#v", specs)
	}
}

func TestResolvedParametersUsesManifestDefaults(t *testing.T) {
	m := &Manifest{
		Parameters: []Parameter{
			{Name: "GO_VERSION", Type: "string", Default: "1.22", Required: true},
			{Name: "RUN_TESTS", Type: "boolean", Default: true},
		},
	}
	resolved := m.ResolvedParameters()
	if resolved["GO_VERSION"] != "1.22" || resolved["RUN_TESTS"] != true {
		t.Fatalf("unexpected parameters: %#v", resolved)
	}
}
