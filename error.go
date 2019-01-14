package logf

import (
	"fmt"
	"time"
)

/*
	[F1, F2, F3]
        |    \-------------------\
        |                        |
	[F1, F2, F3, F4, F5]         |
	                             |
	                    [F1, F2, F3, F6, F7]
	                             |
	                    [F1, F2, F3, F6, F7, F8]
----------------------------------------------------------------------------
	[F1, F2, F3, F4, F5], [F1, F2, F3, F6, F7], [F1, F2, F3, F6, F7, F8] ->
	[F1, F2, F3, F4, F5, F6, F7, F8]
*/

// ErrorEncoder is the function type to encode the given error.
type ErrorEncoder func(string, error, FieldEncoder)

// DefaultErrorEncoder encodes the given error as a set of fields.
//
// A mandatory field with the given key and an optional field with the
// full verbose error message.
func DefaultErrorEncoder(k string, e error, m FieldEncoder) {
	var msg string
	if e == nil {
		msg = "<nil>"
	} else {
		msg = e.Error()
	}
	m.EncodeFieldString(k, msg)

	if _, ok := e.(fmt.Formatter); ok {
		verbose := fmt.Sprintf("%+v", e)
		if verbose != msg {
			m.EncodeFieldString(k+".verbose", verbose)
		}
	}
}

// WrapError wraps the error with fields
// which are used later when logging the error.
func WrapError(err error, fields ...Field) error {
	return &errorWrapper{err, fields[0:len(fields):len(fields)]}
}

// ExtractErrorFields call sink for each field list found in the error if it was created with WrapError.
func ExtractErrorFields(err error, sink func([]Field)) {
	type causer interface {
		Cause() error
	}

	for err != nil {
		switch e := err.(type) {
		case *errorWrapper:
			sink(e.fields)
			err = e.Cause()
		case causer:
			err = e.Cause()
		default:
			err = nil
		}
	}
}

// JoinFields merges two-level list of fields into a single list of fields and removes duplicates.
func JoinFields(result []Field, fields []Field) []Field {
	matching := true
	for i, field := range fields {
		if !matching || len(result) <= i || !result[i].Equal(field) {
			result = append(result, field)
			matching = false
		}
	}

	return result
}

type errorWrapper struct {
	error
	fields []Field
}

func (e *errorWrapper) Cause() error {
	return e.error
}

type prefixingFieldEncoder struct {
	prefix string
	origin FieldEncoder
}

func (e prefixingFieldEncoder) key(k string) string {
	return e.prefix + k
}

func (e prefixingFieldEncoder) EncodeFieldAny(k string, v interface{}) {
	e.origin.EncodeFieldAny(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldBool(k string, v bool) {
	e.origin.EncodeFieldBool(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldInt64(k string, v int64) {
	e.origin.EncodeFieldInt64(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldInt32(k string, v int32) {
	e.origin.EncodeFieldInt32(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldInt16(k string, v int16) {
	e.origin.EncodeFieldInt16(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldInt8(k string, v int8) {
	e.origin.EncodeFieldInt8(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldUint64(k string, v uint64) {
	e.origin.EncodeFieldUint64(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldUint32(k string, v uint32) {
	e.origin.EncodeFieldUint32(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldUint16(k string, v uint16) {
	e.origin.EncodeFieldUint16(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldUint8(k string, v uint8) {
	e.origin.EncodeFieldUint8(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldFloat64(k string, v float64) {
	e.origin.EncodeFieldFloat64(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldFloat32(k string, v float32) {
	e.origin.EncodeFieldFloat32(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldDuration(k string, v time.Duration) {
	e.origin.EncodeFieldDuration(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldError(k string, v error) {
	e.origin.EncodeFieldError(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldTime(k string, v time.Time) {
	e.origin.EncodeFieldTime(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldString(k string, v string) {
	e.origin.EncodeFieldString(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldBytes(k string, v []byte) {
	e.origin.EncodeFieldBytes(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldBools(k string, v []bool) {
	e.origin.EncodeFieldBools(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldInts64(k string, v []int64) {
	e.origin.EncodeFieldInts64(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldInts32(k string, v []int32) {
	e.origin.EncodeFieldInts32(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldInts16(k string, v []int16) {
	e.origin.EncodeFieldInts16(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldInts8(k string, v []int8) {
	e.origin.EncodeFieldInts8(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldUints64(k string, v []uint64) {
	e.origin.EncodeFieldUints64(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldUints32(k string, v []uint32) {
	e.origin.EncodeFieldUints32(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldUints16(k string, v []uint16) {
	e.origin.EncodeFieldUints16(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldUints8(k string, v []uint8) {
	e.origin.EncodeFieldUints8(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldFloats64(k string, v []float64) {
	e.origin.EncodeFieldFloats64(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldFloats32(k string, v []float32) {
	e.origin.EncodeFieldFloats32(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldDurations(k string, v []time.Duration) {
	e.origin.EncodeFieldDurations(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldArray(k string, v ArrayEncoder) {
	e.origin.EncodeFieldArray(e.key(k), v)
}

func (e prefixingFieldEncoder) EncodeFieldObject(k string, v ObjectEncoder) {
	e.origin.EncodeFieldObject(e.key(k), v)
}
