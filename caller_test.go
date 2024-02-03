package logf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryCallerFileWithPackage(t *testing.T) {
	cases := []struct {
		caller CallerInfo
		golden string
	}{
		{
			caller: CallerInfo{0, "/a/b/c/d.go", 66},
			golden: "c/d.go",
		},
		{
			caller: CallerInfo{0, "c/d.go", 66},
			golden: "c/d.go",
		},
		{
			caller: CallerInfo{0, "d.go", 66},
			golden: "d.go",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.golden, c.caller.FileWithPackage())
	}
}

func TestEntryCaller(t *testing.T) {
	caller := NewEntryCaller(0)

	assert.NotEqual(t, 0, caller.PC)
	info, ok := caller.Resolve()
	assert.True(t, ok)
	assert.True(t, info.Line > 0 && info.Line < 1000)
	assert.Equal(t, "logf/caller_test.go", info.FileWithPackage())
	assert.Contains(t, info.File, "/logf/caller_test.go")
}

func TestShortCallerEncoder(t *testing.T) {
	enc := testTypeEncoder{}
	caller := CallerInfo{0, "/a/b/c/d.go", 66}
	ShortCallerEncoder(caller, &enc)

	assert.EqualValues(t, "c/d.go:66", enc.result)
}

func TestFullCallerEncoder(t *testing.T) {
	enc := testTypeEncoder{}
	caller := CallerInfo{0, "/a/b/c/d.go", 66}
	FullCallerEncoder(caller, &enc)

	assert.EqualValues(t, "/a/b/c/d.go:66", enc.result)
}
