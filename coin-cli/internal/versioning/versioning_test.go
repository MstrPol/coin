package versioning

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func initRepo(t *testing.T) (dir string, restore func()) {
	t.Helper()
	dir = t.TempDir()
	oldDir, _ := os.Getwd()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "ci@coin.local")
	run("config", "user.name", "Coin CI")

	f := filepath.Join(dir, "README.md")
	_ = os.WriteFile(f, []byte("# test"), 0644)
	run("add", ".")
	run("commit", "-m", "init")

	_ = os.Chdir(dir)
	restore = func() { _ = os.Chdir(oldDir) }
	return dir, restore
}

func gitTagIn(t *testing.T, dir, tag string) {
	t.Helper()
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git tag %s: %v\n%s", tag, err, out)
	}
}

func checkoutBranch(t *testing.T, dir, branch string) {
	t.Helper()
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b %s: %v\n%s", branch, err, out)
	}
}

func addCommit(t *testing.T, dir, msg string) {
	t.Helper()
	f := filepath.Join(dir, msg+".txt")
	_ = os.WriteFile(f, []byte(msg), 0644)
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("add", ".")
	run("commit", "-m", msg)
}

// ── BranchJiraID ─────────────────────────────────────────────────────────────

func TestBranchJiraID(t *testing.T) {
	cases := []struct{ in, want string }{
		{"release/PROJ-404", "PROJ-404"},
		{"feature/PROJ-101", "PROJ-101"},
		{"feature/PROJ-101-login", "PROJ-101"},
		{"bugfix/ABC-202-fix-null", "ABC-202"},
		{"main", "main"},
		{"master", "master"},
		{"detached", "detached"},
	}
	for _, c := range cases {
		if got := BranchJiraID(c.in); got != c.want {
			t.Errorf("BranchJiraID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── CurrentVersion ────────────────────────────────────────────────────────────

func TestCurrentVersion_NoTags(t *testing.T) {
	_, restore := initRepo(t)
	defer restore()

	v, err := CurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != "0.0.1" {
		t.Errorf("CurrentVersion() = %q, want 0.0.1", v)
	}
}

func TestCurrentVersion_NoRepo(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	v, err := CurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != "0.0.1" {
		t.Errorf("CurrentVersion() = %q, want 0.0.1", v)
	}
}

func TestCurrentVersion_HeadTagged(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	gitTagIn(t, dir, "v1.5.0-PROJ-404-rc-2")

	v, err := CurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != "1.5.0-PROJ-404-rc-2" {
		t.Errorf("CurrentVersion() = %q, want 1.5.0-PROJ-404-rc-2", v)
	}
}

func TestCurrentVersion_LatestTag(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	gitTagIn(t, dir, "v1.5.0-PROJ-404-rc-1")
	addCommit(t, dir, "work")
	gitTagIn(t, dir, "v1.5.0-PROJ-404-rc-2")
	addCommit(t, dir, "more-work")
	// HEAD is not tagged; latest tag is rc-2

	v, err := CurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != "1.5.0-PROJ-404-rc-2" {
		t.Errorf("CurrentVersion() = %q, want 1.5.0-PROJ-404-rc-2", v)
	}
}

// ── NextVersionTag ────────────────────────────────────────────────────────────

func TestNextVersionTag_NewSeriesSnapshot(t *testing.T) {
	_, restore := initRepo(t)
	defer restore()
	// No tags → base = 0.0.0, bump patch → 0.0.1, snapshot-1

	tag, err := NextVersionTag("PROJ-101", "patch", "snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if tag != "v0.0.1-PROJ-101-snapshot-1" {
		t.Errorf("NextVersionTag = %q, want v0.0.1-PROJ-101-snapshot-1", tag)
	}
}

func TestNextVersionTag_NewSeriesMinor(t *testing.T) {
	_, restore := initRepo(t)
	defer restore()

	tag, err := NextVersionTag("PROJ-101", "minor", "snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if tag != "v0.1.0-PROJ-101-snapshot-1" {
		t.Errorf("NextVersionTag = %q, want v0.1.0-PROJ-101-snapshot-1", tag)
	}
}

func TestNextVersionTag_ContinueSnapshotSeries(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	gitTagIn(t, dir, "v0.0.1-PROJ-101-snapshot-1")
	gitTagIn(t, dir, "v0.0.1-PROJ-101-snapshot-2")

	tag, err := NextVersionTag("PROJ-101", "patch", "snapshot")
	if err != nil {
		t.Fatal(err)
	}
	// Series v0.0.1-PROJ-101-snapshot-* exists with max N=2 → continue with N=3
	if tag != "v0.0.1-PROJ-101-snapshot-3" {
		t.Errorf("NextVersionTag = %q, want v0.0.1-PROJ-101-snapshot-3", tag)
	}
}

func TestNextVersionTag_NewRCSeries(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	// Already have a 1.5.0 snapshot series; first RC on release branch
	gitTagIn(t, dir, "v1.5.0-PROJ-404-snapshot-3")

	tag, err := NextVersionTag("PROJ-404", "patch", "rc")
	if err != nil {
		t.Fatal(err)
	}
	// No rc series for PROJ-404 yet → bump patch from 1.5.0 → 1.5.1, rc-1
	if tag != "v1.5.1-PROJ-404-rc-1" {
		t.Errorf("NextVersionTag = %q, want v1.5.1-PROJ-404-rc-1", tag)
	}
}

func TestNextVersionTag_ContinueRCSeries(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	gitTagIn(t, dir, "v1.5.0-PROJ-404-rc-1")
	gitTagIn(t, dir, "v1.5.0-PROJ-404-rc-2")

	// PSI iteration — bump level is irrelevant, series continues
	tag, err := NextVersionTag("PROJ-404", "patch", "rc")
	if err != nil {
		t.Fatal(err)
	}
	if tag != "v1.5.0-PROJ-404-rc-3" {
		t.Errorf("NextVersionTag = %q, want v1.5.0-PROJ-404-rc-3", tag)
	}
}

func TestNextVersionTag_RCNotAffectedBySnapshotSeries(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	// snapshot series for PROJ-404, no RC yet
	gitTagIn(t, dir, "v0.0.1-PROJ-404-snapshot-5")

	tag, err := NextVersionTag("PROJ-404", "minor", "rc")
	if err != nil {
		t.Fatal(err)
	}
	// latest base = 0.0.1, bump minor → 0.1.0, rc-1
	if tag != "v0.1.0-PROJ-404-rc-1" {
		t.Errorf("NextVersionTag = %q, want v0.1.0-PROJ-404-rc-1", tag)
	}
}

func TestNextVersionTag_OtherJiraIDNotAffected(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	gitTagIn(t, dir, "v1.5.0-PROJ-999-rc-5")

	tag, err := NextVersionTag("PROJ-404", "patch", "rc")
	if err != nil {
		t.Fatal(err)
	}
	// PROJ-999 should not affect PROJ-404 counter
	if tag != "v1.5.1-PROJ-404-rc-1" {
		t.Errorf("NextVersionTag = %q, want v1.5.1-PROJ-404-rc-1", tag)
	}
}

// ── bump (pure function) ──────────────────────────────────────────────────────

func TestBump(t *testing.T) {
	cases := []struct {
		in, level, want string
	}{
		{"1.0.0", "patch", "1.0.1"},
		{"1.4.3", "patch", "1.4.4"},
		{"1.0.0", "minor", "1.1.0"},
		{"1.4.3", "minor", "1.5.0"},
		{"1.0.0", "major", "2.0.0"},
		{"1.4.3", "major", "2.0.0"},
		{"0.0.0", "patch", "0.0.1"},
		{"0.0.0", "minor", "0.1.0"},
	}
	for _, c := range cases {
		got, err := bump(c.in, c.level)
		if err != nil {
			t.Errorf("bump(%q, %q): %v", c.in, c.level, err)
			continue
		}
		if got != c.want {
			t.Errorf("bump(%q, %q) = %q, want %q", c.in, c.level, got, c.want)
		}
	}
}

func TestBump_Invalid(t *testing.T) {
	for _, in := range []string{"1.0", "v1.0.0", "latest", ""} {
		if _, err := bump(in, "patch"); err == nil {
			t.Errorf("bump(%q): expected error", in)
		}
	}
}

// ── LatestReleaseTag ──────────────────────────────────────────────────────────

func TestLatestReleaseTag_NoTags(t *testing.T) {
	_, restore := initRepo(t)
	defer restore()

	got, err := LatestReleaseTag("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestLatestReleaseTag_PicksLatest(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	for _, tag := range []string{"v1.0.0-PROJ-100-rc-1", "v1.1.0-PROJ-200-rc-1", "v0.9.0-PROJ-50-rc-3"} {
		gitTagIn(t, dir, tag)
	}

	got, err := LatestReleaseTag("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "v1.1.0-PROJ-200-rc-1" {
		t.Errorf("LatestReleaseTag = %q, want v1.1.0-PROJ-200-rc-1", got)
	}
}

func TestLatestReleaseTag_ExcludeBase(t *testing.T) {
	dir, restore := initRepo(t)
	defer restore()
	for _, tag := range []string{"v1.4.0-PROJ-300-rc-2", "v1.5.0-PROJ-404-rc-1", "v1.5.0-PROJ-404-rc-3"} {
		gitTagIn(t, dir, tag)
	}

	got, err := LatestReleaseTag("1.5.0")
	if err != nil {
		t.Fatal(err)
	}
	if got != "v1.4.0-PROJ-300-rc-2" {
		t.Errorf("LatestReleaseTag(exclude=1.5.0) = %q, want v1.4.0-PROJ-300-rc-2", got)
	}
}

// ── safeName ─────────────────────────────────────────────────────────────────

func TestSafeName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"PROJ-101", "proj-101"},
		{"Feature/My Cool Branch", "feature-my-cool-branch"},
		{"main", "main"},
		{"---foo---", "foo"},
	}
	for _, c := range cases {
		if got := safeName(c.in); got != c.want {
			t.Errorf("safeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── dockerTag ────────────────────────────────────────────────────────────────

func TestDockerTag(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1.2.3", "1.2.3"},
		{"1.5.0-PROJ-404-rc-2", "1.5.0-PROJ-404-rc-2"},
		{"0.0.0-PROJ-101-snapshot-3", "0.0.0-PROJ-101-snapshot-3"},
	}
	for _, c := range cases {
		if got := dockerTag(c.in); !strings.EqualFold(got, c.want) {
			t.Errorf("dockerTag(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
