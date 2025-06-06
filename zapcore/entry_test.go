// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zapcore

import (
	"sync"
	"testing"

	"github.com/toujourser/zap/internal/exit"

	"github.com/stretchr/testify/assert"
)

func assertGoexit(t *testing.T, f func()) {
	var finished bool
	recovered := make(chan interface{})
	go func() {
		defer func() {
			recovered <- recover()
		}()

		f()
		finished = true
	}()

	assert.Nil(t, <-recovered, "Goexit should cause recover to return nil")
	assert.False(t, finished, "Goroutine should not finish after Goexit")
}

func TestPutNilEntry(t *testing.T) {
	// Pooling nil entries defeats the purpose.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			putCheckedEntry(nil)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			ce := getCheckedEntry()
			assert.NotNil(t, ce, "Expected only non-nil CheckedEntries in pool.")
			assert.False(t, ce.dirty, "Unexpected dirty bit set.")
			assert.Nil(t, ce.ErrorOutput, "Non-nil ErrorOutput.")
			assert.Nil(t, ce.after, "Unexpected terminal behavior.")
			assert.Equal(t, 0, len(ce.cores), "Expected empty slice of cores.")
			assert.True(t, cap(ce.cores) > 0, "Expected pooled CheckedEntries to pre-allocate slice of Cores.")
		}
	}()

	wg.Wait()
}

func TestEntryCaller(t *testing.T) {
	tests := []struct {
		caller EntryCaller
		full   string
		short  string
	}{
		{
			caller: NewEntryCaller(100, "/path/to/foo.go", 42, false),
			full:   "undefined",
			short:  "undefined",
		},
		{
			caller: NewEntryCaller(100, "/path/to/foo.go", 42, true),
			full:   "/path/to/foo.go:42",
			short:  "to/foo.go:42",
		},
		{
			caller: NewEntryCaller(100, "to/foo.go", 42, true),
			full:   "to/foo.go:42",
			short:  "to/foo.go:42",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.full, tt.caller.String(), "Unexpected string from EntryCaller.")
		assert.Equal(t, tt.full, tt.caller.FullPath(), "Unexpected FullPath from EntryCaller.")
		assert.Equal(t, tt.short, tt.caller.TrimmedPath(), "Unexpected TrimmedPath from EntryCaller.")
	}
}

func TestCheckedEntryWrite(t *testing.T) {
	t.Run("nil is safe", func(t *testing.T) {
		var ce *CheckedEntry
		assert.NotPanics(t, func() { ce.Write() }, "Unexpected panic writing nil CheckedEntry.")
	})

	t.Run("WriteThenPanic", func(t *testing.T) {
		var ce *CheckedEntry
		ce = ce.After(Entry{}, WriteThenPanic)
		assert.Panics(t, func() { ce.Write() }, "Expected to panic when WriteThenPanic is set.")
	})

	t.Run("WriteThenGoexit", func(t *testing.T) {
		var ce *CheckedEntry
		ce = ce.After(Entry{}, WriteThenGoexit)
		assertGoexit(t, func() { ce.Write() })
	})

	t.Run("WriteThenFatal", func(t *testing.T) {
		var ce *CheckedEntry
		ce = ce.After(Entry{}, WriteThenFatal)
		stub := exit.WithStub(func() {
			ce.Write()
		})
		assert.True(t, stub.Exited, "Expected to exit when WriteThenFatal is set.")
		assert.Equal(t, 1, stub.Code, "Expected to exit when WriteThenFatal is set.")
	})

	t.Run("After", func(t *testing.T) {
		var ce *CheckedEntry
		hook := &customHook{}
		ce = ce.After(Entry{}, hook)
		ce.Write()
		assert.True(t, hook.called, "Expected to call custom action after Write.")
	})
}

type customHook struct {
	called bool
}

func (c *customHook) OnWrite(_ *CheckedEntry, _ []Field) {
	c.called = true
}
