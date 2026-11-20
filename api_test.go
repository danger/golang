package danger_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	danger "github.com/danger/golang"
)

func TestResults(t *testing.T) {
	d := danger.New()
	r, err := d.Results()
	require.Nil(t, err)
	require.Equal(t, `{"fails":[],"warnings":[],"messages":[],"markdowns":[]}`, r)

	d.Message("test", "", 0)
	r, err = d.Results()
	require.Nil(t, err)
	require.Equal(t, `{"fails":[],"warnings":[],"messages":[{"message":"test"}],"markdowns":[]}`, r)
}

func TestResultsAllTypes(t *testing.T) {
	d := danger.New()
	d.Message("a message", "", 0)
	d.Warn("a warning", "api.go", 12)
	d.Fail("a failure", "types.go", 0)
	d.Markdown("## markdown", "README.md", 3)
	d.Message("another message", "main.go", 1)

	r, err := d.Results()
	require.Nil(t, err)
	// This is the shape danger-js reads back from the runner: file and line
	// are omitted when unset, and results keep the order they were added in.
	require.JSONEq(t, `{
		"fails": [{"message":"a failure","file":"types.go"}],
		"warnings": [{"message":"a warning","file":"api.go","line":12}],
		"messages": [{"message":"a message"},{"message":"another message","file":"main.go","line":1}],
		"markdowns": [{"message":"## markdown","file":"README.md","line":3}]
	}`, r)
}
