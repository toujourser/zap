// Copyright (c) 2016-2023 Uber Technologies, Inc.
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

// Package zapfield provides experimental zap.Field helpers whose APIs may be unstable.
package zapfield

import (
	"github.com/toujourser/zap"
	"github.com/toujourser/zap/zapcore"
)

// Str constructs a field with the given string-like key and value.
func Str[K ~string, V ~string](k K, v V) zap.Field {
	return zap.String(string(k), string(v))
}

type stringArray[T ~string] []T

func (a stringArray[T]) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for i := range a {
		enc.AppendString(string(a[i]))
	}
	return nil
}

// Strs constructs a field that carries a slice of string-like values.
func Strs[K ~string, V ~[]S, S ~string](k K, v V) zap.Field {
	return zap.Array(string(k), stringArray[S](v))
}
