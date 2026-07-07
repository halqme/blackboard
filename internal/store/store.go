package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type ProjectConfig struct {
	Version int `json:"version"`
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	Context struct {
		Files []string `json:"files"`
	} `json:"context"`
	Commands map[string]string `json:"commands"`
}

type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Stage     string `json:"stage"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Artifact struct {
	ID              string `json:"id"`
	TaskID          string `json:"task_id"`
	Stage           string `json:"stage"`
	Kind            string `json:"kind"`
	Version         int    `json:"version"`
	Status          string `json:"status"`
	BlobHash        string `json:"blob_hash"`
	BasedOnRevision int    `json:"based_on_revision"`
	CreatedAt       string `json:"created_at"`
}

type State struct {
	ProjectID string     `json:"project_id"`
	Revision  int        `json:"revision"`
	Tasks     []Task     `json:"tasks"`
	Artifacts []Artifact `json:"artifacts"`
}

type Store struct {
	Home      string
	ProjectID string
	Root      string
}

func DefaultHome() string {
	if v := os.Getenv("BLACKBOARD_HOME"); v != "" {
		return v
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return ".blackboard-home"
	}
	return filepath.Join(h, ".local", "share", "blackboard")
}

func New(projectID string) Store {
	home := DefaultHome()
	root := filepath.Join(home, "projects", projectID)
	return Store{Home: home, ProjectID: projectID, Root: root}
}

func (s Store) dbPath() string { return filepath.Join(s.Root, "state.db") }

func (s Store) openDB() (*sql.DB, error) {
	return sql.Open("sqlite", s.dbPath())
}

func (s Store) Ensure() error {
	for _, p := range []string{s.Root, filepath.Join(s.Root, "blobs"), filepath.Join(s.Root, "rendered"), filepath.Join(s.Root, "worktrees")} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			return err
		}
	}
	existed := true
	if _, err := os.Stat(s.dbPath()); os.IsNotExist(err) {
		existed = false
	}
	db, err := s.openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := s.initSchema(db); err != nil {
		return err
	}
	if !existed {
		return s.Save(State{ProjectID: s.ProjectID, Revision: 1, Tasks: []Task{}, Artifacts: []Artifact{}})
	}
	return nil
}

func (s Store) initSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			stage TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS artifacts (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			stage TEXT NOT NULL,
			kind TEXT NOT NULL,
			version INTEGER NOT NULL,
			status TEXT NOT NULL,
			blob_hash TEXT NOT NULL,
			based_on_revision INTEGER NOT NULL,
			created_at TEXT NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s Store) Load() (State, error) {
	db, err := s.openDB()
	if err != nil {
		return State{}, err
	}
	defer db.Close()
	if err := s.initSchema(db); err != nil {
		return State{}, err
	}
	st := State{ProjectID: s.ProjectID}
	var rev string
	if err := db.QueryRow(`SELECT value FROM meta WHERE key='revision'`).Scan(&rev); err != nil {
		return State{}, err
	}
	n, err := strconv.Atoi(rev)
	if err != nil {
		return State{}, err
	}
	st.Revision = n
	rows, err := db.Query(`SELECT id, title, stage, status, created_at, updated_at FROM tasks ORDER BY created_at, id`)
	if err != nil {
		return State{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Stage, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return State{}, err
		}
		st.Tasks = append(st.Tasks, t)
	}
	rows2, err := db.Query(`SELECT id, task_id, stage, kind, version, status, blob_hash, based_on_revision, created_at FROM artifacts ORDER BY created_at, id`)
	if err != nil {
		return State{}, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var a Artifact
		if err := rows2.Scan(&a.ID, &a.TaskID, &a.Stage, &a.Kind, &a.Version, &a.Status, &a.BlobHash, &a.BasedOnRevision, &a.CreatedAt); err != nil {
			return State{}, err
		}
		st.Artifacts = append(st.Artifacts, a)
	}
	return st, nil
}

func (s Store) Save(st State) error {
	db, err := s.openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := s.initSchema(db); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO meta(key, value)
		VALUES ('project_id', ?), ('revision', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, st.ProjectID, fmt.Sprintf("%d", st.Revision)); err != nil {
		return err
	}

	taskIDs := make([]string, 0, len(st.Tasks))
	for _, t := range st.Tasks {
		taskIDs = append(taskIDs, t.ID)
		if _, err := tx.Exec(`
			INSERT INTO tasks(id, title, stage, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				title = excluded.title,
				stage = excluded.stage,
				status = excluded.status,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at
		`, t.ID, t.Title, t.Stage, t.Status, t.CreatedAt, t.UpdatedAt); err != nil {
			return err
		}
	}
	if err := deleteMissingRows(tx, "tasks", taskIDs); err != nil {
		return err
	}

	artifactIDs := make([]string, 0, len(st.Artifacts))
	for _, a := range st.Artifacts {
		artifactIDs = append(artifactIDs, a.ID)
		if _, err := tx.Exec(`
			INSERT INTO artifacts(id, task_id, stage, kind, version, status, blob_hash, based_on_revision, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				task_id = excluded.task_id,
				stage = excluded.stage,
				kind = excluded.kind,
				version = excluded.version,
				status = excluded.status,
				blob_hash = excluded.blob_hash,
				based_on_revision = excluded.based_on_revision,
				created_at = excluded.created_at
		`, a.ID, a.TaskID, a.Stage, a.Kind, a.Version, a.Status, a.BlobHash, a.BasedOnRevision, a.CreatedAt); err != nil {
			return err
		}
	}
	if err := deleteMissingRows(tx, "artifacts", artifactIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func deleteMissingRows(tx *sql.Tx, table string, ids []string) error {
	if len(ids) == 0 {
		_, err := tx.Exec("DELETE FROM " + table)
		return err
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE id NOT IN (%s)", table, placeholders), args...)
	return err
}

func (s Store) PutBlob(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(h.Sum(nil))
	if _, err := f.Seek(0, 0); err != nil {
		return "", err
	}
	dst := filepath.Join(s.Root, "blobs", hash)
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		out, err := os.Create(dst)
		if err != nil {
			return "", err
		}
		defer out.Close()
		if _, err := io.Copy(out, f); err != nil {
			return "", err
		}
	}
	return hash, nil
}

func (s Store) ReadBlob(hash string) (string, error) {
	b, err := os.ReadFile(filepath.Join(s.Root, "blobs", hash))
	return string(b), err
}

func Now() string { return time.Now().UTC().Format(time.RFC3339) }

func NewID(prefix string) string {
	n := time.Now().UnixNano()
	return fmt.Sprintf("%s-%x", prefix, n)
}

func ParseRevision(s string) (int, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
