package presets

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	mfA   = "/tmp/projectA/Makefile"
	mfB   = "/tmp/projectB/Makefile"
	tgBld = "build"
	tgDep = "deploy"
)

// newTempManager returns a Manager backed by a path inside t.TempDir().
// Tests use this everywhere so they can mutate and Save() without
// touching the real cache directory.
func newTempManager(t *testing.T) *Manager {
	t.Helper()
	path := filepath.Join(t.TempDir(), "presets.json")
	mgr, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	return mgr
}

func TestLoadFrom_MissingFileReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	mgr, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
	if mgr.Entries == nil {
		t.Error("expected Entries to be initialized")
	}
	if got := mgr.Count(mfA, tgBld); got != 0 {
		t.Errorf("expected 0 presets, got %d", got)
	}
}

func TestLoadFrom_CorruptJSONReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}
	mgr, _ := LoadFrom(path)
	if mgr == nil || mgr.Entries == nil {
		t.Fatal("expected usable manager with empty Entries on corrupt JSON")
	}
	// And we should be able to write a fresh file on top of it.
	mgr.Upsert(mfA, tgBld, Preset{Name: "p1"})
	if err := mgr.Save(); err != nil {
		t.Fatalf("save after corrupt load: %v", err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	mgr, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	mgr.Upsert(mfA, tgBld, Preset{
		Name:  "fast",
		Env:   map[string]string{"FOO": "1"},
		Flags: []string{"-j4"},
	})
	mgr.Upsert(mfA, tgDep, Preset{
		Name:  "prod",
		Env:   map[string]string{"ENV": "prod"},
		Flags: nil,
	})
	if !mgr.MarkUsed(mfA, tgBld, "fast") {
		t.Fatal("MarkUsed returned false")
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	mgr2, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("re-load: %v", err)
	}

	got, ok := mgr2.Get(mfA, tgBld, "fast")
	if !ok {
		t.Fatal("preset 'fast' missing after reload")
	}
	if !reflect.DeepEqual(got.Env, map[string]string{"FOO": "1"}) {
		t.Errorf("env mismatch after reload: %v", got.Env)
	}
	if !reflect.DeepEqual(got.Flags, []string{"-j4"}) {
		t.Errorf("flags mismatch after reload: %v", got.Flags)
	}
	lu, ok := mgr2.LastUsed(mfA, tgBld)
	if !ok || lu.Name != "fast" {
		t.Errorf("expected LastUsed='fast', got %+v ok=%v", lu, ok)
	}

	// Counts are independent across targets and Makefiles.
	if got := mgr2.Count(mfA, tgBld); got != 1 {
		t.Errorf("Count(build)=%d, want 1", got)
	}
	if got := mgr2.Count(mfA, tgDep); got != 1 {
		t.Errorf("Count(deploy)=%d, want 1", got)
	}
	if got := mgr2.Count(mfB, tgBld); got != 0 {
		t.Errorf("Count(projectB/build)=%d, want 0", got)
	}
}

func TestUpsert_CreatesThenUpdates(t *testing.T) {
	mgr := newTempManager(t)

	created := mgr.Upsert(mfA, tgBld, Preset{Name: "p", Env: map[string]string{"A": "1"}})
	if !created {
		t.Error("first Upsert should return created=true")
	}
	got, _ := mgr.Get(mfA, tgBld, "p")
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Error("CreatedAt and UpdatedAt should be set on insert")
	}
	origCreated := got.CreatedAt

	// Second upsert with same name → update, CreatedAt preserved.
	created = mgr.Upsert(mfA, tgBld, Preset{Name: "p", Env: map[string]string{"A": "2"}})
	if created {
		t.Error("second Upsert should return created=false")
	}
	got, _ = mgr.Get(mfA, tgBld, "p")
	if got.Env["A"] != "2" {
		t.Errorf("Env not updated: %v", got.Env)
	}
	if !got.CreatedAt.Equal(origCreated) {
		t.Errorf("CreatedAt mutated on update: %v -> %v", origCreated, got.CreatedAt)
	}
	if !got.UpdatedAt.After(origCreated) && !got.UpdatedAt.Equal(origCreated) {
		// UpdatedAt may be equal in a sub-millisecond run, so just check
		// it isn't earlier than CreatedAt.
		t.Errorf("UpdatedAt regressed: %v < %v", got.UpdatedAt, origCreated)
	}
}

func TestUpsertStrict_RefusesDuplicate(t *testing.T) {
	mgr := newTempManager(t)

	if err := mgr.UpsertStrict(mfA, tgBld, Preset{Name: "p"}); err != nil {
		t.Fatalf("first UpsertStrict: %v", err)
	}
	err := mgr.UpsertStrict(mfA, tgBld, Preset{Name: "p"})
	if !errors.Is(err, ErrPresetExists) {
		t.Errorf("expected ErrPresetExists, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	mgr := newTempManager(t)
	mgr.Upsert(mfA, tgBld, Preset{Name: "p1"})
	mgr.Upsert(mfA, tgBld, Preset{Name: "p2"})
	mgr.MarkUsed(mfA, tgBld, "p1")

	if !mgr.Delete(mfA, tgBld, "p1") {
		t.Error("Delete should return true for existing preset")
	}
	if mgr.Delete(mfA, tgBld, "p1") {
		t.Error("Delete should return false for missing preset")
	}
	// Deleting the last-used preset clears the pointer.
	if _, ok := mgr.LastUsed(mfA, tgBld); ok {
		t.Error("LastUsed should be empty after deleting the last-used preset")
	}
	// The other preset still exists.
	if !mgr.Exists(mfA, tgBld, "p2") {
		t.Error("Delete removed the wrong preset")
	}
}

func TestList_SortsLastUsedFirst(t *testing.T) {
	mgr := newTempManager(t)
	mgr.Upsert(mfA, tgBld, Preset{Name: "old"})
	mgr.Upsert(mfA, tgBld, Preset{Name: "new"})
	mgr.MarkUsed(mfA, tgBld, "old")

	got := mgr.List(mfA, tgBld)
	if len(got) != 2 {
		t.Fatalf("expected 2 presets, got %d", len(got))
	}
	if got[0].Name != "old" {
		t.Errorf("expected last-used 'old' first, got %q", got[0].Name)
	}
	if got[1].Name != "new" {
		t.Errorf("expected 'new' second, got %q", got[1].Name)
	}
}

func TestMarkUsed_MissingPresetReturnsFalse(t *testing.T) {
	mgr := newTempManager(t)
	if mgr.MarkUsed(mfA, tgBld, "no-such-preset") {
		t.Error("MarkUsed should return false for unknown preset")
	}
	if _, ok := mgr.LastUsed(mfA, tgBld); ok {
		t.Error("LastUsed should remain empty after failed MarkUsed")
	}
}

func TestEmptyEnvAndFlags_NormalizedToNil(t *testing.T) {
	mgr := newTempManager(t)
	mgr.Upsert(mfA, tgBld, Preset{
		Name:  "blank",
		Env:   map[string]string{},
		Flags: []string{},
	})
	got, _ := mgr.Get(mfA, tgBld, "blank")
	if got.Env != nil {
		t.Errorf("empty Env should normalize to nil, got %v", got.Env)
	}
	if got.Flags != nil {
		t.Errorf("empty Flags should normalize to nil, got %v", got.Flags)
	}
}

func TestSaveAtomic_NoStalePresetsOnTempFailure(t *testing.T) {
	// We can't easily simulate a write failure portably, but we can at
	// least assert the .tmp file is cleaned up on success.
	path := filepath.Join(t.TempDir(), "presets.json")
	mgr, _ := LoadFrom(path)
	mgr.Upsert(mfA, tgBld, Preset{Name: "p"})
	if err := mgr.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("expected temp file to be removed after Save, stat err=%v", err)
	}
}

func TestSave_FailsWithoutPath(t *testing.T) {
	mgr := NewEmpty()
	if err := mgr.Save(); err == nil {
		t.Error("Save should fail when path is unset")
	}
}
