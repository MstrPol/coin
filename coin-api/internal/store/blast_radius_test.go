package store

import "testing"

func TestComputeBlastRadius(t *testing.T) {
	versions := []string{"1.0.0", "1.0.0", "1.0.1", "0.9.0"}
	got := computeBlastRadius("go-app", "1.0.2", versions)

	if got.TotalOnGP != 4 {
		t.Fatalf("total %d", got.TotalOnGP)
	}
	if got.OnThisVersion != 0 {
		t.Fatalf("onThisVersion %d", got.OnThisVersion)
	}
	if got.OnOtherVersions != 4 {
		t.Fatalf("onOtherVersions %d", got.OnOtherVersions)
	}
	if got.OnOlderVersions != 4 {
		t.Fatalf("onOlderVersions %d", got.OnOlderVersions)
	}

	got2 := computeBlastRadius("go-app", "1.0.0", versions)
	if got2.OnThisVersion != 2 {
		t.Fatalf("onThisVersion %d", got2.OnThisVersion)
	}
	if got2.OnOlderVersions != 1 {
		t.Fatalf("onOlderVersions %d", got2.OnOlderVersions)
	}
}
