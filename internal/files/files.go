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

func getSnapshotDir() (string, error) {
	// NOTE: maybe this could be configurable?
	// Storing snapshots in root may be desirable in some cases
	snapshotDir := "__snapshots__"
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return "", err
	}

	return snapshotDir, nil
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
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".snap.new") {
			name := strings.TrimSuffix(entry.Name(), ".snap.new")
			newSnapshots = append(newSnapshots, name)
		}
	}

	return newSnapshots, nil
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

func RejectSnapshot(snapTitle string) error {
	filePath, err := getSnapshotPath(snapTitle, "new")
	if err != nil {
		return err
	}

	return os.Remove(filePath)
}
