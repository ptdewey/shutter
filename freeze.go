package freeze

import (
	"fmt"
	"reflect"
)

const version = "0.1.0"

func SnapString(t testingT, content string) {
	t.Helper()
	snap(t, content)
}

func Snap(t testingT, values ...any) {
	t.Helper()
	content := formatValues(values...)
	snap(t, content)
}

func SnapWithTitle(t testingT, title string, values ...any) {
	t.Helper()
	content := formatValues(values...)
	snapWithTitle(t, title, content)
}

func snap(t testingT, content string) {
	t.Helper()
	testName := t.Name()
	snapWithTitle(t, testName, content)
}

func snapWithTitle(t testingT, title string, content string) {
	t.Helper()

	snapshot := &Snapshot{
		Version:  version,
		TestName: title,
		Content:  content,
	}

	accepted, err := readAccepted(title)
	if err == nil {
		if accepted.Content == content {
			return
		}

		if err := SaveSnapshot(snapshot, "new"); err != nil {
			t.Error("failed to save snapshot:", err)
			return
		}

		fmt.Println(DiffSnapshotBox(accepted, snapshot))
		t.Error("snapshot mismatch - run 'freeze review' to update")
		return
	}

	if err := SaveSnapshot(snapshot, "new"); err != nil {
		t.Error("failed to save snapshot:", err)
		return
	}

	fmt.Println(NewSnapshotBox(snapshot))
	t.Error("new snapshot created - run 'freeze review' to accept")
}

func formatValues(values ...any) string {
	if len(values) == 0 {
		return ""
	}

	if len(values) == 1 {
		return formatValue(values[0])
	}

	var result string
	for i, v := range values {
		if i > 0 {
			result += "\n"
		}
		result += formatValue(v)
	}
	return result
}

// TODO: improve this
func formatValue(v any) string {
	if v == nil {
		return "<nil>"
	}

	if formattable, ok := v.(interface{ Format() string }); ok {
		return formattable.Format()
	}

	if stringer, ok := v.(interface{ String() string }); ok {
		return stringer.String()
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.String:
		return v.(string)
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
		// TODO: make this better probably
		return fmt.Sprintf("%#v", v)
	default:
		return fmt.Sprint(v)
	}
}
