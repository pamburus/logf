package logf

import (
	"sync/atomic"
	"time"
)

// NewLogger returns a new Logger with a given Level and EntryWriter.
func NewLogger(level LevelCheckerGetter, w EntryWriter) *Logger {
	return &Logger{
		level: level.LevelChecker(),
		id:    atomic.AddInt32(&nextID, 1),
		w:     w,
	}
}

// NewDisabledLogger return a new Logger that logs nothing as fast as
// possible.
func NewDisabledLogger() *Logger {
	return NewLogger(
		LevelCheckerGetterFunc(func() LevelChecker {
			return func(Level) bool {
				return false
			}
		}), nil)
}

var defaultDisabledLogger = NewDisabledLogger()

// DisabledLogger returns a default instance of a Logger that logs nothing
// as fast as possible.
func DisabledLogger() *Logger {
	return defaultDisabledLogger
}

// LogFunc allows to log a message with a bound level.
type LogFunc func(string, ...Field)

// Logger is the fast, asynchronous, structured logger.
//
// The Logger wraps EntryWriter to check logging level and provide a bit of
// syntactic sugar.
type Logger struct {
	level LevelChecker
	id    int32
	w     EntryWriter

	fields    []Field
	name      string
	addCaller bool
}

// Enabled returns true if logging a message at the specified level is enabled.
func (l *Logger) Enabled(lvl Level) bool {
	return l.level(lvl)
}

// AtLevel calls the given fn if logging a message at the specified level
// is enabled, passing a LogFunc with the bound level.
func (l *Logger) AtLevel(lvl Level, fn func(LogFunc)) {
	if !l.level(lvl) {
		return
	}

	fn(func(text string, fs ...Field) {
		l.log(lvl, 1, text, fs)
	})
}

// WithLevel returns a new logger with the given additional level checker.
func (l *Logger) WithLevel(level LevelCheckerGetter) *Logger {
	cc := l.clone()
	cc.level = func(lvl Level) bool {
		return level.LevelChecker()(lvl) && l.level(lvl)
	}

	return cc
}

// WithName returns a new Logger adding the given name to the calling one.
// Name separator is a period.
//
// Loggers have no name by default.
func (l *Logger) WithName(n string) *Logger {
	if n == "" {
		return l
	}

	cc := l.clone()
	if cc.name == "" {
		cc.name = n
	} else {
		cc.name += "." + n
	}

	return cc
}

// WithCaller returns a new Logger that adds a special annotation parameters
// to each logging message, such as the filename and line number of a caller.
func (l *Logger) WithCaller() *Logger {
	cc := l.clone()
	cc.addCaller = true

	return cc
}

// WithCallerSkip returns a new Logger with increased number of skipped
// frames. It's usable to build a custom wrapper for the Logger.
func (l *Logger) WithCallerSkip(skip int) PreparedLogger {
	return PreparedLogger{l: l, callerSkip: skip}
}

// With returns a new Logger with the given additional fields.
func (l *Logger) With(fs ...Field) *Logger {
	// This code attempts to archive optimum performance with minimum
	// allocations count. Do not change it unless the following benchmarks
	// will show a better performance:
	// - BenchmarkAccumulateFields
	// - BenchmarkAccumulateFieldsWithAccumulatedFields

	cc := l.fork()
	if len(l.fields) == 0 {
		// The fastest way. Use passed 'fs' as is.
		cc.fields = fs
	} else {
		// The less efficient path forces us to copy parent's fields.
		c := make([]Field, 0, len(l.fields)+len(fs))
		c = append(c, l.fields...)
		c = append(c, fs...)

		cc.fields = c
	}

	for i := range cc.fields[len(l.fields):] {
		snapshotField(&cc.fields[i])
	}

	return cc
}

// Log logs a message with the given level, text and optional fields.
func (l *Logger) Log(lvl Level, text string, fs ...Field) {
	l.log(lvl, 1, text, fs)
}

// Debug logs a debug message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Debug(text string, fs ...Field) {
	l.log(LevelDebug, 1, text, fs)
}

// Info logs an info message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Info(text string, fs ...Field) {
	l.log(LevelInfo, 1, text, fs)
}

// Warn logs a warning message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Warn(text string, fs ...Field) {
	l.log(LevelWarn, 1, text, fs)
}

// Error logs an error message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Error(text string, fs ...Field) {
	l.log(LevelError, 1, text, fs)
}

// WithCallerPC returns a new PreparedLogger with the given program counter.
func (l *Logger) WithCallerPC(pc uintptr) PreparedLogger {
	return PreparedLogger{l: l, pc: pc}
}

// WithTime returns a new PreparedLogger with the given timestamp.
func (l *Logger) WithTime(ts time.Time) PreparedLogger {
	return PreparedLogger{l: l, ts: ts}
}

func (l *Logger) log(lvl Level, callerSkip int, text string, fs []Field) {
	l.logCustomized(lvl, callerSkip+1, 0, time.Time{}, text, fs)
}

func (l *Logger) logCustomized(lvl Level, callerSkip int, pc uintptr, ts time.Time, text string, fs []Field) {
	if !l.level(lvl) {
		return
	}

	if ts == (time.Time{}) {
		ts = time.Now()
	}

	// Snapshot non-const fields.
	for i := range fs {
		snapshotField(&fs[i])
	}

	e := Entry{l.id, l.name, l.fields, fs, lvl, ts, text, EntryCaller{}}
	if l.addCaller {
		if pc != 0 {
			e.Caller = NewEntryCallerWithPC(pc)
		} else {
			e.Caller = NewEntryCaller(callerSkip + 1)
		}
	}

	l.w.WriteEntry(e)
}

func (l Logger) clone() *Logger {
	return &l
}

func (l Logger) fork() *Logger {
	l.id = atomic.AddInt32(&nextID, 1)

	return &l
}

type PreparedLogger struct {
	l          *Logger
	callerSkip int
	pc         uintptr
	ts         time.Time
}

// Log logs a message with the given level, text and optional fields.
func (l PreparedLogger) Log(lvl Level, text string, fs ...Field) {
	l.log(lvl, 1, text, fs)
}

// Debug logs a debug message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l PreparedLogger) Debug(text string, fs ...Field) {
	l.log(LevelDebug, 1, text, fs)
}

// Info logs an info message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l PreparedLogger) Info(text string, fs ...Field) {
	l.log(LevelInfo, 1, text, fs)
}

// Warn logs a warning message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l PreparedLogger) Warn(text string, fs ...Field) {
	l.log(LevelWarn, 1, text, fs)
}

// Error logs an error message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l PreparedLogger) Error(text string, fs ...Field) {
	l.log(LevelError, 1, text, fs)
}

// WithCallerPC returns a new PreparedLogger with the given program counter.
func (l PreparedLogger) WithCallerPC(pc uintptr) PreparedLogger {
	l.pc = pc

	return l
}

// WithTime returns a new PreparedLogger with the given timestamp.
func (l PreparedLogger) WithTime(ts time.Time) PreparedLogger {
	l.ts = ts

	return l
}

// WithCallerSkip returns a new PreparedLogger with increased number of skipped frames.
func (l PreparedLogger) WithCallerSkip(callerSkip int) PreparedLogger {
	l.callerSkip += callerSkip

	return l
}

func (l PreparedLogger) log(lvl Level, callerSkip int, text string, fs []Field) {
	l.l.logCustomized(lvl, l.callerSkip+callerSkip+1, l.pc, l.ts, text, fs)
}

var nextID int32
