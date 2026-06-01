package versioning

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ── чистые функции ──────────────────────────────────────────────────────────

func TestSafeName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"feature/PROJ-101", "feature-proj-101"},
		{"Feature/My Cool Branch", "feature-my-cool-branch"},
		{"release/PROJ-404", "release-proj-404"},
		{"main", "main"},
		{"---foo---", "foo"},
		{"UPPER_CASE", "upper-case"},
		{"already-safe.v2", "already-safe.v2"},
	}
	for _, c := range cases {
		got := safeName(c.in)
		if got != c.want {
			t.Errorf("safeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDockerTag(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"1.2.3", "1.2.3"},
		{"0.0.0-main.7+abc1234", "0.0.0-main.7-abc1234"},
		{"0.0.0-feature-proj-101.3+def5678", "0.0.0-feature-proj-101.3-def5678"},
		{"1.5.0-rc.2", "1.5.0-rc.2"},
	}
	for _, c := range cases {
		got := dockerTag(c.in)
		if got != c.want {
			t.Errorf("dockerTag(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── тесты с git-репозиторием ─────────────────────────────────────────────────

// initRepo создаёт временный git-репозиторий и переключает CWD в него.
// Возвращает restore-функцию — обязательно вызывать через defer.
func initRepo(t *testing.T) (dir string, restore func()) {
	t.Helper()

	dir = t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "ci@coin.local")
	run("config", "user.name", "Coin CI")

	// начальный коммит
	f := filepath.Join(dir, "README.md")
	if err := os.WriteFile(f, []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "init")

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	restore = func() {
		_ = os.Chdir(oldDir)
	}
	return dir, restore
}

func TestCompute_NoRepo(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	r, err := Compute("v")
	if err != nil {
		t.Fatal(err)
	}
	if r.Source != "local" {
		t.Errorf("source = %q, want local", r.Source)
	}
	if !strings.HasPrefix(r.Version, "0.0.0-local.") {
		t.Errorf("version = %q, want 0.0.0-local.*", r.Version)
	}
}

func TestCompute_Branch(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()

	// Создаём реальную ветку — ветка берётся из git HEAD, не из BRANCH_NAME
	cmd := exec.Command("git", "checkout", "-b", "feature/PROJ-101")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b: %v\n%s", err, out)
	}

	t.Setenv("BUILD_NUMBER", "42")

	r, err := Compute("v")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(r.Version, "0.0.0-feature-proj-101.42+") {
		t.Errorf("version = %q, expected prefix 0.0.0-feature-proj-101.42+", r.Version)
	}
	if !strings.HasPrefix(r.Source, "branch:feature/PROJ-101:") {
		t.Errorf("source = %q", r.Source)
	}
	// DockerTag не должен содержать '+'
	if strings.Contains(r.ImageTag, "+") {
		t.Errorf("imageTag %q contains '+'", r.ImageTag)
	}
}

func TestCompute_ReleaseBranch(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()

	cmd := exec.Command("git", "checkout", "-b", "release/PROJ-404")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b: %v\n%s", err, out)
	}

	t.Setenv("BUILD_NUMBER", "7")

	r, err := Compute("v")
	if err != nil {
		t.Fatal(err)
	}
	// код: safeName(strings.TrimPrefix("release/PROJ-404","release/")) → "proj-404"
	if !strings.HasPrefix(r.Version, "0.0.0-rc.proj-404.7+") {
		t.Errorf("version = %q, expected prefix 0.0.0-rc.proj-404.7+", r.Version)
	}
	if r.Source != "release-branch:release/PROJ-404" {
		t.Errorf("source = %q", r.Source)
	}
}

func TestCompute_ReleaseTag(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()

	// Ставим тег на HEAD
	cmd := exec.Command("git", "tag", "v1.5.0")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git tag: %v\n%s", err, out)
	}

	r, err := Compute("v")
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != "1.5.0" {
		t.Errorf("version = %q, want 1.5.0", r.Version)
	}
	if r.ImageTag != "1.5.0" {
		t.Errorf("imageTag = %q, want 1.5.0", r.ImageTag)
	}
	if r.Source != "tag:v1.5.0" {
		t.Errorf("source = %q, want tag:v1.5.0", r.Source)
	}
}

func TestCompute_RCTag(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()

	cmd := exec.Command("git", "tag", "v1.5.0-rc.2")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git tag: %v\n%s", err, out)
	}

	r, err := Compute("v")
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != "1.5.0-rc.2" {
		t.Errorf("version = %q, want 1.5.0-rc.2", r.Version)
	}
	if r.Source != "tag:v1.5.0-rc.2" {
		t.Errorf("source = %q", r.Source)
	}
}

func TestCompute_NonReleaseTagIgnored(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()

	// Тег без соответствующего формата — должен игнорироваться
	cmd := exec.Command("git", "tag", "latest")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git tag: %v\n%s", err, out)
	}

	t.Setenv("BRANCH_NAME", "main")
	r, err := Compute("v")
	if err != nil {
		t.Fatal(err)
	}
	// Должен вернуть snapshot, а не тег
	if !strings.HasPrefix(r.Version, "0.0.0-main.") {
		t.Errorf("version = %q, should be snapshot", r.Version)
	}
}

func TestNextRCNumber_NoExisting(t *testing.T) {
	_, restore := initRepo(t)
	defer restore()

	n, err := NextRCNumber("v", "1.5.0")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("NextRCNumber = %d, want 1", n)
	}
}

func TestNextRCNumber_WithExisting(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()

	for _, tag := range []string{"v1.5.0-rc.1", "v1.5.0-rc.2"} {
		cmd := exec.Command("git", "tag", tag)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git tag %s: %v\n%s", tag, err, out)
		}
	}

	n, err := NextRCNumber("v", "1.5.0")
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("NextRCNumber = %d, want 3", n)
	}
}

func TestNextRCNumber_DifferentBaseIgnored(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()

	// Теги другой базовой версии не должны влиять на счётчик
	cmd := exec.Command("git", "tag", "v1.4.0-rc.5")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git tag: %v\n%s", err, out)
	}

	n, err := NextRCNumber("v", "1.5.0")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("NextRCNumber = %d, want 1 (should ignore v1.4.0-rc.5)", n)
	}
}
