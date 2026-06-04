// Package releasenotes генерирует payload для POST /v1/rn (QGM API).
//
// Алгоритм:
//  1. Открывает git-репозиторий в текущей директории.
//  2. Собирает коммиты в диапазоне [from..to] (to по умолчанию — HEAD).
//  3. Из сообщений коммитов извлекает Jira-тикеты (smart commits).
//  4. Формирует map contributors: issue → []ContributorDTO.
//  5. Заполняет buildInfo из переменных окружения Jenkins.
//  6. Возвращает *Payload, готовый для JSON-сериализации и отправки в QGM.
package releasenotes

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// jiraRe находит Jira-тикеты вида PROJ-123, ABC-42 в тексте коммита.
var jiraRe = regexp.MustCompile(`\b([A-Z][A-Z0-9]+-\d+)\b`)

// errStop — сигнал для прерывания итерации по коммитам.
var errStop = errors.New("stop")

// --------------------------------------------------------------------------
// Типы данных (соответствуют схемам QGM API: ReleaseNoteApiRequest)
// --------------------------------------------------------------------------

// Payload — тело запроса POST /v1/rn (ReleaseNoteApiRequest).
type Payload struct {
	Repository   string                     `json:"repository"`
	GroupID      string                     `json:"groupId"`
	ArtifactID   string                     `json:"artifactId"`
	Version      string                     `json:"version"`
	ReleaseNotes []NoteDTO                  `json:"releaseNotes"`
	ReleaseLink  string                     `json:"releaseLink,omitempty"`
	CodeNotes    []CodeNoteDTO              `json:"codeNotes"`
	BuildInfo    *BuildInfoDTO              `json:"buildInfo,omitempty"`
	Meta         []MetaDTO                  `json:"meta"`
	Links        []LinkDTO                  `json:"links"`
	Contributors map[string][]ContributorDTO `json:"contributors"`
	Content      map[string]interface{}     `json:"content"`
}

// NoteDTO — элемент releaseNotes: один Jira-тикет с описанием.
type NoteDTO struct {
	Issue   string `json:"issue"`
	Summary string `json:"summary"`
}

// CodeNoteDTO — информация о диапазоне коммитов для одного репозитория.
// From — SHA коммита предыдущего релиза (нижняя граница, эксклюзивно).
// Пустая строка допустима по API-схеме (pattern: ^(|[a-f0-9]{40})$) и означает,
// что данный релиз является первым в истории репозитория.
type CodeNoteDTO struct {
	Commit     string `json:"commit"`
	Repository string `json:"repository"`
	From       string `json:"from"`
}

// BuildInfoDTO — метаданные Jenkins-сборки.
type BuildInfoDTO struct {
	Meta             []MetaDTO `json:"meta"`
	BuildNumber      string    `json:"buildNumber,omitempty"`
	BuildID          string    `json:"buildId,omitempty"`
	BuildDisplayName string    `json:"buildDisplayName,omitempty"`
	BuildTag         string    `json:"buildTag,omitempty"`
	BuildURL         string    `json:"buildUrl,omitempty"`
	JenkinsURL       string    `json:"jenkinsUrl,omitempty"`
	BranchName       string    `json:"branchName,omitempty"`
	JobName          string    `json:"jobName,omitempty"`
	IPAddress        string    `json:"ipAddress,omitempty"`
	StartTimeMillis  int64     `json:"startTimeInMillis,omitempty"`
}

// MetaDTO — произвольный ключ-значение в meta-блоке.
type MetaDTO struct {
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	Display     string `json:"display,omitempty"`
	Description string `json:"description,omitempty"`
}

// LinkDTO — внешняя ссылка (issue tracker, CI-система и др.).
type LinkDTO struct {
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	Display     string `json:"display,omitempty"`
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain,omitempty"`
}

// ContributorDTO — участник разработки, приписанный к Jira-тикету.
type ContributorDTO struct {
	UserName string `json:"userName,omitempty"`
	Email    string `json:"email,omitempty"`
	Login    string `json:"login,omitempty"`
}

// --------------------------------------------------------------------------
// Опции генерации
// --------------------------------------------------------------------------

// Options — входные параметры для Generate().
type Options struct {
	// Поля артефакта (из .coin/config.yaml).
	Repository string
	GroupID    string
	ArtifactID string
	Version    string

	// From — тег или SHA нижней границы диапазона (не включается).
	// Пусто → все коммиты до HEAD (первый релиз).
	From string

	// RepoDir — путь к git-репозиторию (для тестов). Пусто → cwd, с поиском .git вверх.
	RepoDir string

	// CodeRepository — URL git-репозитория для поля codeNotes.repository.
	// Если пусто — берётся автоматически из remote «origin».
	CodeRepository string

	// Опциональная ссылка на Jira-задачу «Release 2.0».
	ReleaseLink string
}

// --------------------------------------------------------------------------
// Основная функция
// --------------------------------------------------------------------------

// Generate собирает Payload из git-истории проекта.
func Generate(opts Options) (*Payload, error) {
	repoPath := opts.RepoDir
	if repoPath == "" {
		repoPath = "."
	}
	openOpts := &git.PlainOpenOptions{DetectDotGit: opts.RepoDir == ""}
	repo, err := git.PlainOpenWithOptions(repoPath, openOpts)
	if err != nil {
		return nil, fmt.Errorf("не git-репозиторий: %w", err)
	}

	// Верхняя граница — всегда HEAD.
	toSHA, err := resolveRef(repo, "")
	if err != nil {
		return nil, fmt.Errorf("HEAD: %w", err)
	}

	// rangeFromSHA — исключающая нижняя граница для фильтрации коммитов.
	// codeNoteFromSHA — SHA для поля codeNotes.from (всегда конкретный коммит).
	rangeFromSHA := ""
	codeNoteFromSHA := ""

	if opts.From != "" {
		// Не первый релиз: from-тег есть, разрезолвить в SHA.
		rangeFromSHA, err = resolveRef(repo, opts.From)
		if err != nil {
			return nil, fmt.Errorf("разрешение from-тега %q: %w", opts.From, err)
		}
		codeNoteFromSHA = rangeFromSHA
	} else {
		// Первый релиз: нижней границы нет, но codeNotes.from = SHA первого коммита.
		if first, err := firstCommitSHA(repo); err == nil {
			codeNoteFromSHA = first
		}
	}

	commits, err := commitsInRange(repo, rangeFromSHA, toSHA)
	if err != nil {
		return nil, fmt.Errorf("чтение коммитов: %w", err)
	}

	// Парсим smart-commits: тикеты и контрибьюторы.
	issueOrder := []string{}
	issueSummary := map[string]string{}
	contributors := map[string][]ContributorDTO{}

	for _, c := range commits {
		issues := jiraRe.FindAllString(c.Message, -1)
		for _, issue := range issues {
			if _, seen := issueSummary[issue]; !seen {
				issueOrder = append(issueOrder, issue)
				issueSummary[issue] = commitSubject(c.Message)
			}
			contrib := ContributorDTO{
				UserName: c.Author.Name,
				Email:    c.Author.Email,
			}
			if !hasContributor(contributors[issue], contrib.Email) {
				contributors[issue] = append(contributors[issue], contrib)
			}
		}
	}

	// Собираем releaseNotes в порядке первого появления.
	releaseNotes := make([]NoteDTO, 0, len(issueOrder))
	for _, issue := range issueOrder {
		releaseNotes = append(releaseNotes, NoteDTO{Issue: issue, Summary: issueSummary[issue]})
	}

	// URL репозитория: из конфига или автоматически из remote origin.
	codeRepo := opts.CodeRepository
	if codeRepo == "" {
		codeRepo = remoteURL(repo)
	}

	codeNotes := []CodeNoteDTO{
		{Commit: toSHA, Repository: codeRepo, From: codeNoteFromSHA},
	}

	buildInfo := collectBuildInfo()

	meta := []MetaDTO{
		{
			Key:         "coin.version",
			Value:       opts.Version,
			Display:     "Coin Version",
			Description: "Версия артефакта из git-тега Coin",
		},
		{
			Key:     "generated.at",
			Value:   time.Now().UTC().Format(time.RFC3339),
			Display: "Generated At",
		},
	}

	links := []LinkDTO{}

	return &Payload{
		Repository:   opts.Repository,
		GroupID:      opts.GroupID,
		ArtifactID:   opts.ArtifactID,
		Version:      opts.Version,
		ReleaseNotes: releaseNotes,
		ReleaseLink:  opts.ReleaseLink,
		CodeNotes:    codeNotes,
		BuildInfo:    buildInfo,
		Meta:         meta,
		Links:        links,
		Contributors: contributors,
		Content:      map[string]interface{}{},
	}, nil
}

// SaveToFile записывает payload в JSON-файл, создавая родительские директории.
func SaveToFile(payload *Payload, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("создание директории %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("сериализация JSON: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("запись файла %s: %w", path, err)
	}
	return nil
}

// --------------------------------------------------------------------------
// Вспомогательные функции
// --------------------------------------------------------------------------

// resolveRef возвращает SHA коммита для тега, ветки или SHA (или HEAD если ref пусто).
func resolveRef(repo *git.Repository, ref string) (string, error) {
	if ref == "" {
		head, err := repo.Head()
		if err != nil {
			return "", fmt.Errorf("HEAD: %w", err)
		}
		return head.Hash().String(), nil
	}

	// Сначала пробуем как аннотированный тег.
	tagRef, err := repo.Tag(ref)
	if err == nil {
		tagObj, err := repo.TagObject(tagRef.Hash())
		if err == nil {
			return tagObj.Target.String(), nil
		}
		// Лёгкий тег — хеш это уже коммит.
		return tagRef.Hash().String(), nil
	}

	// Пробуем как ссылку (ветка, HEAD и т.д.).
	r, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err == nil {
		return r.String(), nil
	}

	// Принимаем как сырой SHA (минимум 7 символов).
	if len(ref) >= 7 {
		return ref, nil
	}

	return "", fmt.Errorf("не найден ref %q", ref)
}

// commitsInRange возвращает коммиты в диапазоне (fromSHA, toSHA].
// fromSHA пусто — возвращает все коммиты, достижимые из toSHA.
func commitsInRange(repo *git.Repository, fromSHA, toSHA string) ([]*object.Commit, error) {
	toHash := plumbing.NewHash(toSHA)
	toCommit, err := repo.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("коммит %s не найден: %w", toSHA, err)
	}

	iter, err := repo.Log(&git.LogOptions{From: toCommit.Hash})
	if err != nil {
		return nil, err
	}

	var commits []*object.Commit
	iterErr := iter.ForEach(func(c *object.Commit) error {
		if fromSHA != "" && c.Hash.String() == fromSHA {
			return errStop
		}
		commits = append(commits, c)
		return nil
	})
	if iterErr != nil && !errors.Is(iterErr, errStop) {
		return nil, iterErr
	}

	return commits, nil
}

// commitSubject возвращает первую строку сообщения коммита.
func commitSubject(msg string) string {
	return strings.TrimSpace(strings.SplitN(strings.TrimSpace(msg), "\n", 2)[0])
}

// hasContributor проверяет, есть ли контрибьютор с таким email в списке.
func hasContributor(list []ContributorDTO, email string) bool {
	for _, c := range list {
		if c.Email == email {
			return true
		}
	}
	return false
}

// firstCommitSHA возвращает SHA самого старого коммита, достижимого из HEAD.
// Используется как codeNotes.from при отсутствии предыдущего релиза.
func firstCommitSHA(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	iter, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return "", err
	}
	var oldest *object.Commit
	_ = iter.ForEach(func(c *object.Commit) error {
		oldest = c
		return nil
	})
	if oldest == nil {
		return "", fmt.Errorf("нет коммитов")
	}
	return oldest.Hash.String(), nil
}

// remoteURL возвращает URL remote «origin» через go-git (без exec).
// Если remote не настроен — возвращает пустую строку.
func remoteURL(repo *git.Repository) string {
	remote, err := repo.Remote("origin")
	if err != nil || len(remote.Config().URLs) == 0 {
		return ""
	}
	return remote.Config().URLs[0]
}

// collectBuildInfo читает стандартные переменные окружения Jenkins.
func collectBuildInfo() *BuildInfoDTO {
	return &BuildInfoDTO{
		Meta:             []MetaDTO{},
		BuildNumber:      os.Getenv("BUILD_NUMBER"),
		BuildID:          os.Getenv("BUILD_ID"),
		BuildDisplayName: os.Getenv("BUILD_DISPLAY_NAME"),
		BuildTag:         os.Getenv("BUILD_TAG"),
		BuildURL:         os.Getenv("BUILD_URL"),
		JenkinsURL:       os.Getenv("JENKINS_URL"),
		BranchName:       os.Getenv("BRANCH_NAME"),
		JobName:          os.Getenv("JOB_NAME"),
	}
}
