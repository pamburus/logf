package logf

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultErrorEncoderWithPlainError(t *testing.T) {
	e := errors.New("simple error")
	enc := newTestFieldEncoder()
	DefaultErrorEncoder("error", e, enc)

	assert.EqualValues(t, e.Error(), enc.result["error"])
	assert.EqualValues(t, 1, len(enc.result))
}

func TestDefaultErrorEncoderWithVerboseError(t *testing.T) {
	e := &verboseError{"short", "verbose"}
	enc := newTestFieldEncoder()
	DefaultErrorEncoder("error", e, enc)

	assert.EqualValues(t, 2, len(enc.result))
	assert.EqualValues(t, e.short, enc.result["error"])
	assert.EqualValues(t, e.full, enc.result["error.verbose"])
}

func TestDefaultErrorEncoderWithNil(t *testing.T) {
	enc := newTestFieldEncoder()
	DefaultErrorEncoder("error", nil, enc)

	assert.EqualValues(t, 1, len(enc.result))
	assert.EqualValues(t, "<nil>", enc.result["error"])
}

func TestErrorFieldsExtraction(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		sets := [][]Field{}
		ExtractErrorFields(nil, func(fields []Field) {
			sets = append(sets, fields)
		})

		require.Equal(t, 0, len(sets))
	})
	t.Run("ZeroWrap", func(t *testing.T) {
		sets := [][]Field{}
		ExtractErrorFields(errors.New("test error"), func(fields []Field) {
			sets = append(sets, fields)
		})

		require.Equal(t, 0, len(sets))
	})
	t.Run("SingleWrap", func(t *testing.T) {
		sets := [][]Field{}
		fields := []Field{
			String("f1", "f1v"),
			Int("f2", 1),
		}
		err := WrapError(errors.New("test error"), fields...)
		ExtractErrorFields(err, func(fields []Field) {
			sets = append(sets, fields)
		})

		require.Equal(t, 1, len(sets))
		assert.Equal(t, fields, sets[0])
	})
	t.Run("TripleWrap", func(t *testing.T) {
		sets := [][]Field{}
		fields1 := []Field{
			String("f1", "f1v"),
			Int("f2", 1),
		}
		fields2 := append(fields1, String("f3", "f3v"))
		fields3 := append(fields2, Bool("f4", true))
		err := WrapError(errors.New("test error"), fields3...)
		err = WrapError(err, fields2...)
		err = WrapError(err, fields1...)
		ExtractErrorFields(err, func(fields []Field) {
			sets = append(sets, fields)
		})

		require.Equal(t, 3, len(sets))
		assert.Equal(t, fields1, sets[0])
		assert.Equal(t, fields2, sets[1])
		assert.Equal(t, fields3, sets[2])
	})
}

func TestErrorFieldsJoining(t *testing.T) {
	t.Run("Sequential", func(t *testing.T) {
		fields1 := []Field{
			String("f1", "f1v"),
			Int("f2", 1),
		}
		fields2 := append(fields1, String("f3", "f3v"))
		fields3 := append(fields2, Bool("f4", true))
		err := WrapError(errors.New("test error"), fields3...)
		err = WrapError(err, fields2...)
		err = WrapError(err, fields1...)
		result := []Field{}
		ExtractErrorFields(err, func(fields []Field) {
			result = JoinFields(result, fields)
		})

		require.Equal(t, 4, len(result))
		assert.Equal(t, fields1[0].Key, result[0].Key)
		assert.Equal(t, fields1[1].Key, result[1].Key)
		assert.Equal(t, fields2[2].Key, result[2].Key)
		assert.Equal(t, fields3[3].Key, result[3].Key)
	})
	t.Run("Mixed", func(t *testing.T) {
		fields1 := []Field{
			String("f1", "f1v"),
			Int("f2", 1),
		}
		fields2 := append(fields1, String("f3", "f3v"))
		fields3 := append(fields1, Bool("f4", true))
		err := WrapError(errors.New("test error"), fields3...)
		err = WrapError(err, fields2...)
		err = WrapError(err, fields1...)
		result := []Field{}
		ExtractErrorFields(err, func(fields []Field) {
			result = JoinFields(result, fields)
		})

		require.Equal(t, 4, len(result))
		assert.Equal(t, fields1[0].Key, result[0].Key)
		assert.Equal(t, fields1[1].Key, result[1].Key)
		assert.Equal(t, fields2[2].Key, result[2].Key)
		assert.Equal(t, fields3[2].Key, result[3].Key)
	})
}
