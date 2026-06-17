package build

import (
	"strings"
	"testing"
)

func TestPackArgs(t *testing.T) {
	opts := BuildpackOptions{
		Workspace: "/ws",
		Builder:   "paketobuildpacks/builder-jammy-base",
		RunImage:  "gcr.io/distroless/base-debian12",
		CacheRef:  "nexus:8082/coin-cache/app:buildpack",
		Output:    "type=image,name=nexus:8082/coin-docker/app:1.0.0,push=false",
		Env:       []string{"BP_RUN_TESTS=true"},
	}
	args, err := packArgs(opts)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{
		"build nexus:8082/coin-docker/app:1.0.0",
		"--builder paketobuildpacks/builder-jammy-base",
		"--path /ws",
		"--run-image gcr.io/distroless/base-debian12",
		"--env BP_RUN_TESTS=true",
		"--trust-builder",
		"--docker-host inherit",
		"--network host",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %q", want, joined)
		}
	}
	if strings.Contains(joined, "--publish") {
		t.Fatalf("unexpected --publish in %q", joined)
	}
}

func TestPackArgsPublish(t *testing.T) {
	args, err := packArgs(BuildpackOptions{
		Workspace: "/ws",
		Builder:   "paketobuildpacks/builder-jammy-base",
		CacheRef:  "nexus:8082/coin-cache/app:buildpack",
		Output:    "type=image,name=nexus:8082/coin-docker/app:1.0.0,push=true",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(args, " "), "--publish") {
		t.Fatal("expected --publish for push=true output")
	}
	if !strings.Contains(strings.Join(args, " "), "--cache-image nexus:8082/coin-cache/app:buildpack") {
		t.Fatal("expected --cache-image with publish")
	}
}

func TestPackArgsTestDiscardImage(t *testing.T) {
	args, err := packArgs(BuildpackOptions{
		Workspace: "/ws",
		Builder:   "paketobuildpacks/builder-jammy-base",
		Env:       []string{"BP_RUN_TESTS=true"},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "build localhost/coin-buildpack-test:discard") {
		t.Fatalf("expected discard image for test build, got %q", joined)
	}
	if strings.Contains(joined, "--cache-image") {
		t.Fatalf("test build must not use cache image: %q", joined)
	}
}

func TestParsePackImageRef(t *testing.T) {
	ref, pub, err := parsePackImageRef("type=image,name=reg/app:1,push=true")
	if err != nil || ref != "reg/app:1" || !pub {
		t.Fatalf("got ref=%q pub=%v err=%v", ref, pub, err)
	}
}
