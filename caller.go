package logf

import (
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// EntryCaller holds values returned by runtime.Caller.
type EntryCaller struct {
	PC uintptr
}

// Resolve returns the file and line number of the EntryCaller.
func (c EntryCaller) Resolve() (CallerInfo, bool) {
	var info CallerInfo
	if c.PC == 0 {
		return info, false
	}

	var pcs = [1]uintptr{c.PC}
	frame, _ := runtime.CallersFrames(pcs[:]).Next()
	info.PC = frame.PC
	info.File = frame.File
	info.Line = frame.Line

	return info, frame.PC != 0
}

// NewEntryCaller creates an instance of EntryCaller with the given number
// of frames to skip.
func NewEntryCaller(skip int) EntryCaller {
	var pcs [1]uintptr
	runtime.Callers(skip+2, pcs[:])

	return EntryCaller{pcs[0]}
}

// NewEntryCallerWithPC creates an instance of EntryCaller with the given program counter.
func NewEntryCallerWithPC(pc uintptr) EntryCaller {
	return EntryCaller{pc}
}

type CallerInfo struct {
	PC   uintptr
	File string
	Line int
}

// FileWithPackage cuts a package name and a file name from EntryCaller.File.
func (c CallerInfo) FileWithPackage() string {

	// As for os-specific path separator battle here, my opinion coincides
	// with the last comment here https://github.com/golang/go/issues/3335.
	//
	// Go team should simply document the current behavior of always using
	// '/' in stack frame data. That's the way it's been implemented for
	// years, and packages like github.com/go-stack/stack that have been
	// stable for years expect it. Changing the behavior in a future version
	// of Go will break working code without a clearly documented benefit.
	// Documenting the behavior will help avoid new code from making the
	// wrong assumptions.

	found := strings.LastIndexByte(c.File, '/')
	if found == -1 {
		return c.File
	}
	found = strings.LastIndexByte(c.File[:found], '/')
	if found == -1 {
		return c.File
	}

	return c.File[found+1:]
}

// CallerEncoder is the function type to encode the given EntryCaller.
type CallerEncoder func(CallerInfo, TypeEncoder)

// ShortCallerEncoder encodes the given EntryCaller using it's FileWithPackage
// function.
func ShortCallerEncoder(c CallerInfo, m TypeEncoder) {
	var callerBuf [64]byte
	var b []byte
	b = callerBuf[:0]
	b = append(b, c.FileWithPackage()...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(c.Line), 10)

	m.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

// FullCallerEncoder encodes the given EntryCaller using a full file path.
func FullCallerEncoder(c CallerInfo, m TypeEncoder) {
	var callerBuf [256]byte
	var b []byte
	b = callerBuf[:0]
	b = append(b, c.File...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(c.Line), 10)

	m.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}
