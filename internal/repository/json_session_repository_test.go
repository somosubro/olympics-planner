package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONSessionRepository_RejectsObjectWrapper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")
	content := `{"sessions": [{"id": "x"}]}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := JSONSessionRepository{Path: path}
	_, err := repo.GetAll()
	if err == nil {
		t.Fatal("expected error for object wrapper root")
	}
	if err != ErrSessionsFileNotJSONArray {
		t.Fatalf("expected ErrSessionsFileNotJSONArray, got %v", err)
	}
}

func TestJSONSessionRepository_AcceptsJSONArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")
	content := `[{"id":"a","sport":"Tennis","sessionCode":"T1","date":"2028-07-15","dayOfWeek":"Saturday","startTime":"10:00","venue":"V"}]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := JSONSessionRepository{Path: path}
	sessions, err := repo.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "a" {
		t.Fatalf("got %#v", sessions)
	}
}
