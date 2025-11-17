package files

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Snapshot struct {
	Version  string
	TestName string
	Content  string
}

func (s *Snapshot) Serialize() string {
	header := fmt.Sprintf("---\nversion: %s\ntest_name: %s\n---\n", s.Version, s.TestName)
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

	for _, line := range strings.Split(header, "\n") {
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
		case "version":
			snap.Version = value
		case "test_name":
			snap.TestName = value
		}
	}

	return snap, nil
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := cwd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("go.mod not found")
		}
		current = parent
	}
}

func getSnapshotDir() (string, error) {
	root, err := findProjectRoot()
	if err != nil {
		return "", err
	}

	snapshotDir := filepath.Join(root, "__snapshots__")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return "", err
	}

	return snapshotDir, nil
}

func SnapshotFileName(testName string) string {
	var result strings.Builder
	for i, r := range testName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	s := result.String()
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	return s
}

func SaveSnapshot(snap *Snapshot, state string) error {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return err
	}

	fileName := SnapshotFileName(snap.TestName) + "." + state
	filePath := filepath.Join(snapshotDir, fileName)

	return os.WriteFile(filePath, []byte(snap.Serialize()), 0644)
}

func ReadSnapshot(testName string, state string) (*Snapshot, error) {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return nil, err
	}

	fileName := SnapshotFileName(testName) + "." + state
	filePath := filepath.Join(snapshotDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Deserialize(string(data))
}

func ReadAccepted(testName string) (*Snapshot, error) {
	return ReadSnapshot(testName, "accepted")
}

func ReadNew(testName string) (*Snapshot, error) {
	return ReadSnapshot(testName, "new")
}

func ListNewSnapshots() ([]string, error) {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return nil, err
	}

	var newSnapshots []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".new") {
			name := strings.TrimSuffix(entry.Name(), ".new")
			newSnapshots = append(newSnapshots, name)
		}
	}

	return newSnapshots, nil
}

func AcceptSnapshot(testName string) error {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return err
	}

	fileName := SnapshotFileName(testName)
	newPath := filepath.Join(snapshotDir, fileName+".new")
	acceptedPath := filepath.Join(snapshotDir, fileName+".accepted")

	data, err := os.ReadFile(newPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(acceptedPath, data, 0644); err != nil {
		return err
	}

	return os.Remove(newPath)
}

func RejectSnapshot(testName string) error {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return err
	}

	fileName := SnapshotFileName(testName) + ".new"
	filePath := filepath.Join(snapshotDir, fileName)

	return os.Remove(filePath)
}
