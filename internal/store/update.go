package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

var ErrRevisionMismatch = errors.New("revision mismatch")

func (s Store) Update(expectedRevision int, mutate func(*State) error) (State, error) {
	db, err := s.openDB()
	if err != nil {
		return State{}, err
	}
	defer db.Close()
	if err := s.initSchema(db); err != nil {
		return State{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return State{}, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE meta SET value = ? WHERE key = 'revision' AND value = ?`, strconv.Itoa(expectedRevision+1), strconv.Itoa(expectedRevision))
	if err != nil {
		return State{}, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return State{}, err
	}
	if changed != 1 {
		return State{}, ErrRevisionMismatch
	}
	st, err := loadStateTx(tx, s.ProjectID, expectedRevision)
	if err != nil {
		return State{}, err
	}
	if err := mutate(&st); err != nil {
		return State{}, err
	}
	st.ProjectID = s.ProjectID
	st.Revision = expectedRevision + 1
	if err := syncStateTx(tx, st); err != nil {
		return State{}, err
	}
	if err := tx.Commit(); err != nil {
		return State{}, err
	}
	return st, nil
}

func loadStateTx(tx *sql.Tx, projectID string, revision int) (State, error) {
	st := State{ProjectID: projectID, Revision: revision}
	rows, err := tx.Query(`SELECT id, title, stage, status, created_at, updated_at FROM tasks ORDER BY created_at, id`)
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
	if err := rows.Err(); err != nil {
		return State{}, err
	}
	artifacts, err := tx.Query(`SELECT id, task_id, stage, kind, version, status, blob_hash, based_on_revision, created_at FROM artifacts ORDER BY created_at, id`)
	if err != nil {
		return State{}, err
	}
	defer artifacts.Close()
	for artifacts.Next() {
		var a Artifact
		if err := artifacts.Scan(&a.ID, &a.TaskID, &a.Stage, &a.Kind, &a.Version, &a.Status, &a.BlobHash, &a.BasedOnRevision, &a.CreatedAt); err != nil {
			return State{}, err
		}
		st.Artifacts = append(st.Artifacts, a)
	}
	if err := artifacts.Err(); err != nil {
		return State{}, err
	}
	deps, err := tx.Query(`SELECT artifact_id, depends_on_artifact_id FROM artifact_dependencies ORDER BY artifact_id, depends_on_artifact_id`)
	if err != nil {
		return State{}, err
	}
	defer deps.Close()
	byID := make(map[string]*Artifact, len(st.Artifacts))
	for i := range st.Artifacts {
		byID[st.Artifacts[i].ID] = &st.Artifacts[i]
	}
	for deps.Next() {
		var artifactID, dependsOn string
		if err := deps.Scan(&artifactID, &dependsOn); err != nil {
			return State{}, err
		}
		if a := byID[artifactID]; a != nil {
			a.DependsOn = append(a.DependsOn, dependsOn)
		}
	}
	if err := deps.Err(); err != nil {
		return State{}, err
	}
	return st, nil
}

func syncStateTx(tx *sql.Tx, st State) error {
	if _, err := tx.Exec(`INSERT INTO meta(key, value) VALUES ('project_id', ?), ('revision', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, st.ProjectID, fmt.Sprintf("%d", st.Revision)); err != nil {
		return err
	}
	taskIDs := make([]string, 0, len(st.Tasks))
	for _, t := range st.Tasks {
		taskIDs = append(taskIDs, t.ID)
		if _, err := tx.Exec(`INSERT INTO tasks(id, title, stage, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET title=excluded.title, stage=excluded.stage, status=excluded.status, created_at=excluded.created_at, updated_at=excluded.updated_at`, t.ID, t.Title, t.Stage, t.Status, t.CreatedAt, t.UpdatedAt); err != nil {
			return err
		}
	}
	if err := deleteMissingRows(tx, "tasks", taskIDs); err != nil {
		return err
	}
	artifactIDs := make([]string, 0, len(st.Artifacts))
	for _, a := range st.Artifacts {
		artifactIDs = append(artifactIDs, a.ID)
		if _, err := tx.Exec(`INSERT INTO artifacts(id, task_id, stage, kind, version, status, blob_hash, based_on_revision, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET task_id=excluded.task_id, stage=excluded.stage, kind=excluded.kind, version=excluded.version, status=excluded.status, blob_hash=excluded.blob_hash, based_on_revision=excluded.based_on_revision, created_at=excluded.created_at`, a.ID, a.TaskID, a.Stage, a.Kind, a.Version, a.Status, a.BlobHash, a.BasedOnRevision, a.CreatedAt); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM artifact_dependencies`); err != nil {
		return err
	}
	for _, a := range st.Artifacts {
		for _, dep := range a.DependsOn {
			if _, err := tx.Exec(`INSERT INTO artifact_dependencies(artifact_id, depends_on_artifact_id) VALUES (?, ?)`, a.ID, dep); err != nil {
				return err
			}
		}
	}
	return deleteMissingRows(tx, "artifacts", artifactIDs)
}
