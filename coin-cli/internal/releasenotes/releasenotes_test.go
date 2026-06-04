package releasenotes

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testAuthorEmail = "ci@coin.local"
const testAuthorName = "Coin CI"

// ── helpers ───────────────────────────────────────────────────────────────────

func gitEnv() []string {
	return append(os.Environ(),
		"GIT_AUTHOR_NAME="+testAuthorName,
		"GIT_AUTHOR_EMAIL="+testAuthorEmail,
		"GIT_COMMITTER_NAME="+testAuthorName,
		"GIT_COMMITTER_EMAIL="+testAuthorEmail,
	)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", testAuthorEmail)
	runGit(t, dir, "config", "user.name", testAuthorName)
	addCommitIn(t, dir, "init: initial commit", "README.md", "# test")
	return dir
}

func addCommitIn(t *testing.T, dir, msg, file, content string) {
	t.Helper()
	f := filepath.Join(dir, file)
	_ = os.WriteFile(f, []byte(content), 0644)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", msg)
}

func tagIn(t *testing.T, dir, tag string) {
	t.Helper()
	runGit(t, dir, "tag", tag)
}

func headSHAIn(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	cmd.Env = gitEnv()
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func optsInRepo(dir string, opts Options) Options {
	opts.RepoDir = dir
	return opts
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCommitSubject(t *testing.T) {
	cases := []struct {
		msg  string
		want string
	}{
		{"feat(rn): добавить генерацию release notes", "feat(rn): добавить генерацию release notes"},
		{"feat(rn): добавить генерацию release notes\n\nДетали...", "feat(rn): добавить генерацию release notes"},
		{"  trimmed  \n\nmore", "trimmed"},
	}
	for _, tc := range cases {
		if got := commitSubject(tc.msg); got != tc.want {
			t.Errorf("commitSubject(%q) = %q, want %q", tc.msg, got, tc.want)
		}
	}
}

func TestHasContributor(t *testing.T) {
	list := []ContributorDTO{{Email: "a@b.com"}, {Email: "c@d.com"}}
	if !hasContributor(list, "a@b.com") {
		t.Error("expected true for a@b.com")
	}
	if hasContributor(list, "x@y.com") {
		t.Error("expected false for x@y.com")
	}
}

func TestJiraRe(t *testing.T) {
	msg := "feat: PROJ-123 fix issue ABC-1 and MYTEAM-999"
	got := jiraRe.FindAllString(msg, -1)
	if len(got) != 3 {
		t.Fatalf("got %v, want 3 matches", got)
	}
	if got[0] != "PROJ-123" || got[1] != "ABC-1" || got[2] != "MYTEAM-999" {
		t.Errorf("unexpected matches: %v", got)
	}
}

func TestGenerate_NoCommits(t *testing.T) {
	dir := initRepo(t)

	opts := optsInRepo(dir, Options{
		Repository: "Nexus_PROD",
		GroupID:    "com.example",
		ArtifactID: "my-app",
		Version:    "1.0.0-PROJ-100-rc-1",
	})

	payload, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if payload.Repository != "Nexus_PROD" {
		t.Errorf("repository = %q, want Nexus_PROD", payload.Repository)
	}
	if payload.ArtifactID != "my-app" {
		t.Errorf("artifactId = %q, want my-app", payload.ArtifactID)
	}
	if payload.Content == nil {
		t.Error("content должен быть непустым map")
	}
	if len(payload.ReleaseNotes) != 0 {
		t.Errorf("releaseNotes = %v, ожидали пустой список", payload.ReleaseNotes)
	}
}

func TestGenerate_ExtractsJiraIssues(t *testing.T) {
	dir := initRepo(t)

	addCommitIn(t, dir, "feat: PROJ-42 добавить фичу", "feat.txt", "feature")
	addCommitIn(t, dir, "fix: PROJ-43 исправить баг\n\nДетали исправления", "fix.txt", "fix")

	opts := optsInRepo(dir, Options{
		Repository: "Nexus_PROD",
		GroupID:    "com.example",
		ArtifactID: "my-app",
		Version:    "1.0.0-PROJ-100-rc-1",
	})

	payload, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	issueMap := map[string]string{}
	for _, rn := range payload.ReleaseNotes {
		issueMap[rn.Issue] = rn.Summary
	}

	if _, ok := issueMap["PROJ-42"]; !ok {
		t.Error("PROJ-42 не найден в releaseNotes")
	}
	if _, ok := issueMap["PROJ-43"]; !ok {
		t.Error("PROJ-43 не найден в releaseNotes")
	}

	if len(payload.Contributors["PROJ-42"]) == 0 {
		t.Error("contributors для PROJ-42 пустые")
	}
	if payload.Contributors["PROJ-42"][0].Email != testAuthorEmail {
		t.Errorf("email = %q, want %s", payload.Contributors["PROJ-42"][0].Email, testAuthorEmail)
	}
}

func TestGenerate_FromToRange(t *testing.T) {
	dir := initRepo(t)

	addCommitIn(t, dir, "feat: OLD-1 старая фича", "old.txt", "old")
	fromSHA := headSHAIn(t, dir)

	addCommitIn(t, dir, "feat: NEW-1 новая фича", "new1.txt", "new1")
	addCommitIn(t, dir, "fix: NEW-2 исправить", "new2.txt", "new2")

	opts := optsInRepo(dir, Options{
		Repository: "Nexus_PROD",
		GroupID:    "com.example",
		ArtifactID: "my-app",
		Version:    "1.1.0-PROJ-200-rc-1",
		From:       fromSHA,
	})

	payload, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	issueMap := map[string]bool{}
	for _, rn := range payload.ReleaseNotes {
		issueMap[rn.Issue] = true
	}

	if issueMap["OLD-1"] {
		t.Error("OLD-1 не должен быть в payload (до from)")
	}
	if !issueMap["NEW-1"] {
		t.Error("NEW-1 должен быть в payload")
	}
	if !issueMap["NEW-2"] {
		t.Error("NEW-2 должен быть в payload")
	}
}

func TestGenerate_DedupContributors(t *testing.T) {
	dir := initRepo(t)

	addCommitIn(t, dir, "feat: DEDUP-1 первый коммит", "d1.txt", "a")
	addCommitIn(t, dir, "fix: DEDUP-1 второй коммит", "d2.txt", "b")

	opts := optsInRepo(dir, Options{
		Repository: "Nexus_PROD",
		GroupID:    "com.example",
		ArtifactID: "my-app",
		Version:    "1.0.0-DEDUP-1-rc-1",
	})

	payload, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	contribs := payload.Contributors["DEDUP-1"]
	emails := map[string]int{}
	for _, c := range contribs {
		emails[c.Email]++
	}
	if emails[testAuthorEmail] != 1 {
		t.Errorf("%s встречается %d раз, ожидали 1 (дедупликация)", testAuthorEmail, emails[testAuthorEmail])
	}
}

func TestGenerate_MetaHasVersion(t *testing.T) {
	dir := initRepo(t)

	opts := optsInRepo(dir, Options{
		Repository: "Nexus_PROD",
		GroupID:    "com.example",
		ArtifactID: "my-app",
		Version:    "2.3.4-PROJ-99-rc-5",
	})

	payload, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	found := false
	for _, m := range payload.Meta {
		if m.Key == "coin.version" && m.Value == opts.Version {
			found = true
		}
	}
	if !found {
		t.Errorf("meta не содержит coin.version=%q, meta=%v", opts.Version, payload.Meta)
	}
}

func TestCodeNotes_FromAlwaysPresent(t *testing.T) {
	dir := initRepo(t)

	addCommitIn(t, dir, "feat: PROJ-1 фича", "feat.txt", "content")

	opts := optsInRepo(dir, Options{
		Repository: "Nexus_PROD",
		GroupID:    "com.example",
		ArtifactID: "my-app",
		Version:    "1.0.0-PROJ-1-rc-1",
	})

	payload, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if len(payload.CodeNotes) == 0 {
		t.Fatal("codeNotes пустой")
	}

	cn := payload.CodeNotes[0]

	if len(cn.From) != 40 {
		t.Errorf("from = %q, ожидали 40-символьный SHA", cn.From)
	}
	if cn.Commit == cn.From {
		t.Errorf("commit == from (%s): HEAD и первый коммит не должны совпадать", cn.Commit)
	}
	data, err := json.Marshal(cn)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"from"`) {
		t.Errorf("поле from отсутствует в JSON: %s", data)
	}
}

func TestSaveToFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".coin", "temp", "release-notes.json")

	payload := &Payload{
		Repository:   "Nexus_PROD",
		GroupID:      "com.example",
		ArtifactID:   "my-app",
		Version:      "1.0.0",
		ReleaseNotes: []NoteDTO{{Issue: "PROJ-1", Summary: "test"}},
		CodeNotes:    []CodeNoteDTO{},
		Meta:         []MetaDTO{},
		Links:        []LinkDTO{},
		Contributors: map[string][]ContributorDTO{},
		Content:      map[string]interface{}{},
	}

	if err := SaveToFile(payload, path); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if len(data) == 0 {
		t.Error("файл пустой")
	}
	if string(data[:1]) != "{" {
		t.Error("файл не начинается с {")
	}
}
