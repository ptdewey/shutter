package freeze

type Snapshot struct {
	Version  string
	TestName string
	Content  string
}

type Config struct {
	snapshotDir string
	extension   string
}

func Frame(t testingT, vals ...any) {
	t.Helper()
}

func newSnapshot(name, content string, cfg Config) {

}
