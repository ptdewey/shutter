package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Snapshot struct {
	Version  string
	Title    string
	Test     string
	FileName string
	Content  string
}

func (s *Snapshot) Serialize() string {
	header := fmt.Sprintf(
		"---\ntitle: %s\ntest_name: %s\nfile_name: %s\nversion: %s\n---\n",
		s.Title, s.Test, s.FileName, s.Version,
	)
	return header + s.Content
}

func Deserialize(raw string) (*Snapshot, error) {
	parts := strings.SplitN(raw, "---\n", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid snapshot format")
	}

	header := parts[1]
	content := parts[2]

	snap := &Snapshot{
		Content: content,
	}

	for line := range strings.SplitSeq(header, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		kv := strings.SplitN(line, ": ", 2)
		if len(kv) != 2 {
			continue
		}

		key, value := kv[0], kv[1]
		switch key {
		case "title":
			snap.Title = value
		case "test_name":
			snap.Test = value
		case "file_name":
			snap.FileName = value
		case "version":
			snap.Version = value
		}
	}

	return snap, nil
}

// getSnapshotDir finds the nearest __snapshots__ directory relative to the caller,
// creating one if it doesn't exist. This is used when creating new snapshots.
func getSnapshotDir() (string, error) {
	// NOTE: maybe this could be configurable?
	// Storing snapshots in root may be desirable in some cases
	snapshotDir := "__snapshots__"
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return "", err
	}

	return snapshotDir, nil
}

// findAllSnapshotDirs recursively finds all __snapshots__ directories starting from root
func findAllSnapshotDirs(root string) ([]string, error) {
	var snapshotDirs []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common ignore paths
		if info.IsDir() && len(info.Name()) > 0 && info.Name()[0] == '.' {
			return filepath.SkipDir
		}
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		if info.IsDir() && info.Name() == "__snapshots__" {
			snapshotDirs = append(snapshotDirs, path)
		}

		return nil
	})

	return snapshotDirs, err
}

// findProjectRoot finds the root of the project by looking for go.mod
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			// Fall back to current directory
			return cwd, nil
		}
		dir = parent
	}
}

// TODO: make this use the snapshot title rather than the test name
func SnapshotFileName(snapTitle string) string {
	return strings.ReplaceAll(strings.ToLower(snapTitle), " ", "_")
}

// getSnapshotFileName returns the filename for a snapshot based on test name and state
func getSnapshotFileName(snapTitle string, state string) string {
	baseName := SnapshotFileName(snapTitle)
	switch state {
	case "accepted":
		return baseName + ".snap"
	case "new":
		return baseName + ".snap.new"
	default:
		return baseName + "." + state
	}
}

// getSnapshotPath returns the full path for a snapshot file
func getSnapshotPath(snapTitle string, state string) (string, error) {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return "", err
	}

	fileName := getSnapshotFileName(snapTitle, state)
	return filepath.Join(snapshotDir, fileName), nil
}

func SaveSnapshot(snap *Snapshot, state string) error {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return err
	}

	fileName := getSnapshotFileName(snap.Title, state)
	filePath := filepath.Join(snapshotDir, fileName)

	return os.WriteFile(filePath, []byte(snap.Serialize()), 0644)
}

func ReadSnapshot(snapTitle string, state string) (*Snapshot, error) {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return nil, err
	}

	return ReadSnapshotWithDir(snapshotDir, snapTitle, state)
}

// ReadSnapshotFromPath reads a snapshot directly from a full file path
func ReadSnapshotFromPath(filePath string) (*Snapshot, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Deserialize(string(data))
}

// ReadSnapshotWithDir reads a snapshot from a specific directory
func ReadSnapshotWithDir(snapshotDir, snapTitle string, state string) (*Snapshot, error) {
	fileName := getSnapshotFileName(snapTitle, state)
	filePath := filepath.Join(snapshotDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Deserialize(string(data))
}

func ReadAccepted(snapTitle string) (*Snapshot, error) {
	return ReadSnapshot(snapTitle, "snap")
}

func ReadNew(snapTitle string) (*Snapshot, error) {
	return ReadSnapshot(snapTitle, "new")
}

// SnapshotInfo contains metadata about a snapshot file including its full path
type SnapshotInfo struct {
	Title string // The snapshot title (used as identifier)
	Path  string // Full path to the snapshot file
	Dir   string // Directory containing the snapshot
}

func ListNewSnapshots() ([]SnapshotInfo, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, err
	}

	snapshotDirs, err := findAllSnapshotDirs(projectRoot)
	if err != nil {
		return nil, err
	}

	var newSnapshots []SnapshotInfo
	for _, dir := range snapshotDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			// Skip directories we can't read
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".snap.new") {
				name := strings.TrimSuffix(entry.Name(), ".snap.new")
				fullPath := filepath.Join(dir, entry.Name())
				newSnapshots = append(newSnapshots, SnapshotInfo{
					Title: name,
					Path:  fullPath,
					Dir:   dir,
				})
			}
		}
	}

	return newSnapshots, nil
}

// AcceptSnapshotInfo accepts a snapshot using SnapshotInfo
func AcceptSnapshotInfo(info SnapshotInfo) error {
	newPath := info.Path
	acceptedPath := filepath.Join(info.Dir, getSnapshotFileName(info.Title, "accepted"))

	data, err := os.ReadFile(newPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(acceptedPath, data, 0644); err != nil {
		return err
	}

	return os.Remove(newPath)
}

func AcceptSnapshot(snapTitle string) error {
	newPath, err := getSnapshotPath(snapTitle, "new")
	if err != nil {
		return err
	}

	acceptedPath, err := getSnapshotPath(snapTitle, "accepted")
	if err != nil {
		return err
	}

	data, err := os.ReadFile(newPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(acceptedPath, data, 0644); err != nil {
		return err
	}

	return os.Remove(newPath)
}

// RejectSnapshotInfo rejects a snapshot using SnapshotInfo
func RejectSnapshotInfo(info SnapshotInfo) error {
	return os.Remove(info.Path)
}

func RejectSnapshot(snapTitle string) error {
	filePath, err := getSnapshotPath(snapTitle, "new")
	if err != nil {
		return err
	}

	return os.Remove(filePath)
}
