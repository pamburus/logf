package logf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type gateWriter struct {
	active     chan struct{}
	entered    chan string
	release    chan struct{}
	concurrent chan string
}

func newGateWriter() *gateWriter {
	return &gateWriter{
		active:     make(chan struct{}, 1),
		entered:    make(chan string, 8),
		release:    make(chan struct{}, 8),
		concurrent: make(chan string, 8),
	}
}

func (w *gateWriter) enter(op string) {
	select {
	case w.active <- struct{}{}:
	default:
		w.concurrent <- op
	}

	w.entered <- op
	<-w.release
	<-w.active
}

type writeOnlyGateWriter struct{ *gateWriter }

func (w *writeOnlyGateWriter) Write(p []byte) (int, error) {
	w.enter("write")
	return len(p), nil
}

type flushGateWriter struct{ *gateWriter }

func (w *flushGateWriter) Write(p []byte) (int, error) {
	w.enter("write")
	return len(p), nil
}

func (w *flushGateWriter) Flush() error {
	w.enter("flush")
	return nil
}

type syncGateWriter struct {
	*gateWriter
	probed bool
}

func (w *syncGateWriter) Write(p []byte) (int, error) {
	w.enter("write")
	return len(p), nil
}

func (w *syncGateWriter) Sync() error {
	if !w.probed {
		w.probed = true
		return nil
	}
	w.enter("sync")
	return nil
}

func TestWriterFromIOSerializesConcurrentWrites(t *testing.T) {
	dst := &writeOnlyGateWriter{gateWriter: newGateWriter()}
	w := WriterFromIO(dst)

	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		_, err := w.Write([]byte("first"))
		require.NoError(t, err)
	}()

	require.Equal(t, "write", waitOp(t, dst.entered))

	secondStarted := make(chan struct{})
	secondDone := make(chan struct{})
	go func() {
		close(secondStarted)
		defer close(secondDone)
		_, err := w.Write([]byte("second"))
		require.NoError(t, err)
	}()
	<-secondStarted

	assertNoOp(t, dst.entered)
	assertNotDone(t, secondDone)

	dst.release <- struct{}{}
	waitDone(t, firstDone)

	require.Equal(t, "write", waitOp(t, dst.entered))
	dst.release <- struct{}{}
	waitDone(t, secondDone)

	assertNoConcurrentOp(t, dst.concurrent)
}

func TestWriterFromIOSerializesFlushWithWrite(t *testing.T) {
	dst := &flushGateWriter{gateWriter: newGateWriter()}
	w := WriterFromIO(dst)

	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		_, err := w.Write([]byte("write"))
		require.NoError(t, err)
	}()

	require.Equal(t, "write", waitOp(t, dst.entered))

	flushStarted := make(chan struct{})
	flushDone := make(chan struct{})
	go func() {
		close(flushStarted)
		defer close(flushDone)
		require.NoError(t, w.Flush())
	}()
	<-flushStarted

	assertNoOp(t, dst.entered)
	assertNotDone(t, flushDone)

	dst.release <- struct{}{}
	waitDone(t, writeDone)

	require.Equal(t, "flush", waitOp(t, dst.entered))
	dst.release <- struct{}{}
	waitDone(t, flushDone)

	assertNoConcurrentOp(t, dst.concurrent)
}

func TestWriterFromIOSerializesSyncWithWrite(t *testing.T) {
	dst := &syncGateWriter{gateWriter: newGateWriter()}
	w := WriterFromIO(dst)

	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		_, err := w.Write([]byte("write"))
		require.NoError(t, err)
	}()

	require.Equal(t, "write", waitOp(t, dst.entered))

	syncStarted := make(chan struct{})
	syncDone := make(chan struct{})
	go func() {
		close(syncStarted)
		defer close(syncDone)
		require.NoError(t, w.Sync())
	}()
	<-syncStarted

	assertNoOp(t, dst.entered)
	assertNotDone(t, syncDone)

	dst.release <- struct{}{}
	waitDone(t, writeDone)

	require.Equal(t, "sync", waitOp(t, dst.entered))
	dst.release <- struct{}{}
	waitDone(t, syncDone)

	assertNoConcurrentOp(t, dst.concurrent)
}

func waitOp(t *testing.T, ch <-chan string) string {
	t.Helper()
	select {
	case op := <-ch:
		return op
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for writer operation")
		return ""
	}
}

func waitDone(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for goroutine completion")
	}
}

func assertNoOp(t *testing.T, ch <-chan string) {
	t.Helper()
	select {
	case op := <-ch:
		t.Fatalf("unexpected overlapping operation: %s", op)
	case <-time.After(50 * time.Millisecond):
	}
}

func assertNotDone(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
		t.Fatal("operation completed before prior call was released")
	default:
	}
}

func assertNoConcurrentOp(t *testing.T, ch <-chan string) {
	t.Helper()
	select {
	case op := <-ch:
		t.Fatalf("underlying writer was entered concurrently during %s", op)
	default:
	}
}
