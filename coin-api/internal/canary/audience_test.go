package canary

import "testing"

func TestProjectBucketDeterministic(t *testing.T) {
	a := ProjectBucket("demo-go-app")
	b := ProjectBucket("demo-go-app")
	if a != b {
		t.Fatalf("bucket not deterministic: %d vs %d", a, b)
	}
	if a < 0 || a >= 100 {
		t.Fatalf("bucket out of range: %d", a)
	}
}

func TestUseCanaryLine(t *testing.T) {
	if UseCanaryLine("x", "canary", 0, true) != true {
		t.Fatal("canary mode should always use canary")
	}
	if UseCanaryLine("x", "stable", 100, true) != false {
		t.Fatal("stable mode should never use canary")
	}
	if UseCanaryLine("", "default", 100, true) != false {
		t.Fatal("empty project should default to stable")
	}
	if UseCanaryLine("x", "default", 0, true) != false {
		t.Fatal("0% should never use canary for default projects")
	}
}
