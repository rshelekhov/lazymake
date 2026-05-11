// Package presets provides persistent, named run configurations
// (environment variables + make flags) for Makefile targets.
//
// A preset is identified by the tuple (absolute makefile path, target
// name, preset name). Presets are stored in a single JSON file under
// the user cache directory and follow the same graceful-degradation
// pattern as internal/history: a missing or corrupt file results in an
// empty manager and the rest of the app keeps running.
package presets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	presetsFileName = "presets.json"
)

// Preset captures one named configuration for a target.
//
// Env and Flags are omitted from JSON when empty so a "plain" preset
// (e.g. one that just records the intent of running with no extras)
// produces a compact on-disk representation.
type Preset struct {
	Name       string            `json:"name"`
	Env        map[string]string `json:"env,omitempty"`
	Flags      []string          `json:"flags,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	LastUsedAt time.Time         `json:"last_used_at,omitempty"`
}

// TargetPresets is the bucket of presets for a single target. The
// LastUsedPreset pointer is stored alongside the presets so the
// "R: rerun last-used" hotkey can resolve in O(1) without scanning
// LastUsedAt timestamps.
type TargetPresets struct {
	Presets        []Preset `json:"presets"`
	LastUsedPreset string   `json:"last_used_preset,omitempty"`
}

// Manager owns the in-memory copy of presets.json and serializes
// mutations through Save(). Concurrent use is not supported; the TUI
// drives Manager from the single Bubble Tea event loop.
type Manager struct {
	// Entries[absoluteMakefilePath][targetName] -> bucket of presets.
	// We use a pointer to TargetPresets so callers can take a reference,
	// mutate, and have it round-trip through MarshalJSON unchanged.
	Entries map[string]map[string]*TargetPresets `json:"entries"`

	path string // file path; "" means "Save is a no-op"
}

// ErrPresetExists is returned by UpsertStrict when the (makefile,
// target, name) tuple is already taken. The TUI uses this to decide
// when to prompt the user with an "overwrite?" confirmation.
var ErrPresetExists = errors.New("preset with this name already exists")

// Load reads presets.json from the user cache directory. On any I/O
// or parse failure it returns an empty (but writable) Manager along
// with the error, matching the history package's graceful pattern so
// callers can decide whether to log or silently ignore.
func Load() (*Manager, error) {
	path, err := getCachePath()
	if err != nil {
		return newEmpty(), fmt.Errorf("failed to get cache path: %w", err)
	}
	return LoadFrom(path)
}

// LoadFrom reads presets from an explicit path. Exposed primarily so
// tests can point at t.TempDir() without monkey-patching the cache
// directory.
func LoadFrom(path string) (*Manager, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			m := newEmpty()
			m.path = path
			return m, nil
		}
		m := newEmpty()
		m.path = path
		return m, fmt.Errorf("failed to read presets file: %w", err)
	}

	var m Manager
	if err := json.Unmarshal(data, &m); err != nil {
		// Corrupt JSON: keep the user running with an empty store, but
		// surface the problem on stderr so they can investigate.
		_, _ = fmt.Fprintf(os.Stderr, "Warning: corrupt presets file, resetting: %v\n", err)
		m = *newEmpty()
	}

	if m.Entries == nil {
		m.Entries = make(map[string]map[string]*TargetPresets)
	}
	m.path = path
	return &m, nil
}

// NewEmpty returns a Manager backed by no on-disk file. Save() will
// fail until the caller assigns a path or constructs the manager via
// Load/LoadFrom. Used by callers that want a usable manager even when
// Load returns an error.
func NewEmpty() *Manager {
	return newEmpty()
}

func newEmpty() *Manager {
	return &Manager{
		Entries: make(map[string]map[string]*TargetPresets),
	}
}

// Save writes the manager to disk atomically (write-then-rename) so a
// crash mid-write cannot leave a partial file behind. The parent
// directory is created on demand.
func (m *Manager) Save() error {
	if m.path == "" {
		return errors.New("presets path not set")
	}

	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal presets: %w", err)
	}

	// Atomic write: write to a sibling temp file, then rename. Rename
	// is atomic on POSIX and on Windows (same volume) per Go 1.5+.
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("failed to write presets temp file: %w", err)
	}
	if err := os.Rename(tmp, m.path); err != nil {
		// Clean up the temp file on rename failure to avoid leaving
		// clutter behind.
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to commit presets file: %w", err)
	}
	return nil
}

// List returns all presets for (makefilePath, targetName), sorted with
// the last-used preset first, then by UpdatedAt descending so the
// freshest variant always appears near the top.
func (m *Manager) List(makefilePath, targetName string) []Preset {
	bucket := m.bucket(makefilePath, targetName)
	if bucket == nil || len(bucket.Presets) == 0 {
		return nil
	}

	out := make([]Preset, len(bucket.Presets))
	copy(out, bucket.Presets)

	lastUsed := bucket.LastUsedPreset
	sort.SliceStable(out, func(i, j int) bool {
		if lastUsed != "" {
			if out[i].Name == lastUsed && out[j].Name != lastUsed {
				return true
			}
			if out[j].Name == lastUsed && out[i].Name != lastUsed {
				return false
			}
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

// Get returns the preset matching (makefilePath, targetName, name) by
// value, along with a found flag. Returning by value keeps callers
// from accidentally mutating the manager's internal state.
func (m *Manager) Get(makefilePath, targetName, name string) (Preset, bool) {
	bucket := m.bucket(makefilePath, targetName)
	if bucket == nil {
		return Preset{}, false
	}
	for _, p := range bucket.Presets {
		if p.Name == name {
			return p, true
		}
	}
	return Preset{}, false
}

// Exists reports whether (makefilePath, targetName, name) is already
// taken. The TUI uses this to decide between "save new" and "prompt
// for overwrite confirmation".
func (m *Manager) Exists(makefilePath, targetName, name string) bool {
	_, ok := m.Get(makefilePath, targetName, name)
	return ok
}

// Upsert creates the preset if absent, otherwise updates the existing
// one in place. The created return value tells the caller which path
// was taken (useful for status messages). Empty Env/Flags are accepted
// and round-trip as nil.
func (m *Manager) Upsert(makefilePath, targetName string, p Preset) (created bool) {
	now := time.Now()
	p.Name = strings.TrimSpace(p.Name)
	if p.Env != nil && len(p.Env) == 0 {
		p.Env = nil
	}
	if p.Flags != nil && len(p.Flags) == 0 {
		p.Flags = nil
	}

	bucket := m.ensureBucket(makefilePath, targetName)
	for i := range bucket.Presets {
		if bucket.Presets[i].Name == p.Name {
			// Preserve original CreatedAt to keep the audit trail intact.
			bucket.Presets[i].Env = p.Env
			bucket.Presets[i].Flags = p.Flags
			bucket.Presets[i].UpdatedAt = now
			return false
		}
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	bucket.Presets = append(bucket.Presets, p)
	return true
}

// UpsertStrict refuses to overwrite an existing preset. It returns
// ErrPresetExists in that case so the TUI can branch into an overwrite
// confirmation flow.
func (m *Manager) UpsertStrict(makefilePath, targetName string, p Preset) error {
	if m.Exists(makefilePath, targetName, p.Name) {
		return ErrPresetExists
	}
	m.Upsert(makefilePath, targetName, p)
	return nil
}

// Delete removes the named preset and reports whether anything was
// actually deleted. If the deleted preset was the LastUsedPreset, the
// pointer is cleared so a stale name doesn't linger.
func (m *Manager) Delete(makefilePath, targetName, name string) bool {
	bucket := m.bucket(makefilePath, targetName)
	if bucket == nil {
		return false
	}
	for i := range bucket.Presets {
		if bucket.Presets[i].Name == name {
			bucket.Presets = append(bucket.Presets[:i], bucket.Presets[i+1:]...)
			if bucket.LastUsedPreset == name {
				bucket.LastUsedPreset = ""
			}
			return true
		}
	}
	return false
}

// MarkUsed updates LastUsedAt on the preset and pins it as the
// LastUsedPreset for its target so the "R" rerun hotkey can find it.
// Returns false when the preset doesn't exist (caller decides what to
// do — typically nothing).
func (m *Manager) MarkUsed(makefilePath, targetName, name string) bool {
	bucket := m.bucket(makefilePath, targetName)
	if bucket == nil {
		return false
	}
	for i := range bucket.Presets {
		if bucket.Presets[i].Name == name {
			bucket.Presets[i].LastUsedAt = time.Now()
			bucket.LastUsedPreset = name
			return true
		}
	}
	return false
}

// LastUsed returns the last-used preset for (makefilePath, targetName)
// when one is recorded and still exists. The pointer is reset on
// delete, so a non-empty LastUsedPreset always resolves.
func (m *Manager) LastUsed(makefilePath, targetName string) (Preset, bool) {
	bucket := m.bucket(makefilePath, targetName)
	if bucket == nil || bucket.LastUsedPreset == "" {
		return Preset{}, false
	}
	return m.Get(makefilePath, targetName, bucket.LastUsedPreset)
}

// Count returns the number of presets registered for (makefilePath,
// targetName). Cheap query for "N presets available" hints in the UI.
func (m *Manager) Count(makefilePath, targetName string) int {
	bucket := m.bucket(makefilePath, targetName)
	if bucket == nil {
		return 0
	}
	return len(bucket.Presets)
}

// bucket returns the existing bucket or nil — never auto-creates.
// Used by read-only operations.
func (m *Manager) bucket(makefilePath, targetName string) *TargetPresets {
	if m == nil || m.Entries == nil {
		return nil
	}
	byTarget, ok := m.Entries[makefilePath]
	if !ok {
		return nil
	}
	return byTarget[targetName]
}

// ensureBucket returns the existing bucket or creates an empty one.
// Used by write operations.
func (m *Manager) ensureBucket(makefilePath, targetName string) *TargetPresets {
	if m.Entries == nil {
		m.Entries = make(map[string]map[string]*TargetPresets)
	}
	byTarget, ok := m.Entries[makefilePath]
	if !ok {
		byTarget = make(map[string]*TargetPresets)
		m.Entries[makefilePath] = byTarget
	}
	bucket, ok := byTarget[targetName]
	if !ok {
		bucket = &TargetPresets{}
		byTarget[targetName] = bucket
	}
	return bucket
}

// getCachePath returns the platform-appropriate path to presets.json.
// It prefers os.UserCacheDir() and falls back to ~/.cache to stay
// consistent with internal/history/history.go.
func getCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", herr
		}
		cacheDir = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheDir, "lazymake", presetsFileName), nil
}
