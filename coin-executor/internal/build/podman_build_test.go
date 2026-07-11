package build

import "testing"

func TestParseImageOutput(t *testing.T) {
	ref, push, err := parseImageOutput("type=image,name=nexus:8082/coin-docker/app:1.0.0,push=true")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "nexus:8082/coin-docker/app:1.0.0" || !push {
		t.Fatalf("got ref=%q push=%v", ref, push)
	}
}

func TestParseImageOutputNoPush(t *testing.T) {
	ref, push, err := parseImageOutput("type=image,name=localhost/app:dev,push=false")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "localhost/app:dev" || push {
		t.Fatalf("got ref=%q push=%v", ref, push)
	}
}
