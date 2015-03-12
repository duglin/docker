// A generic GO interface copier
// From: https://gist.github.com/hvoecking/10772475
//
// The MIT License (MIT)
//
// Copyright (c) 2014 Heye VÃ¶cking
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

package config

import (
	"reflect"
)

func Copy(obj interface{}) interface{} {
	copy := CopyValue(reflect.ValueOf(obj))

	// Remove the reflection wrapper
	return copy.Interface()
}

func CopyValue(orig reflect.Value) reflect.Value {
	copy := reflect.New(orig.Type()).Elem()
	copyRecursive(copy, orig)
	return copy
}

func copyRecursive(copy, orig reflect.Value) {
	switch orig.Kind() {

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		origValue := orig.Elem()
		// Check if the pointer is nil
		if !origValue.IsValid() {
			return
		}
		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(origValue.Type()))
		// Unwrap the newly created pointer
		copyRecursive(copy.Elem(), origValue)

	// If it is an interface (which is very similar to a pointer), do basically
	// the same as for the pointer. Though a pointer is not the same as an
	// interface so note that we have to call Elem() after creating a new
	// object because otherwise we would end up with an actual pointer
	case reflect.Interface:
		// Get rid of the wrapping interface
		origValue := orig.Elem()
		// Create a new object. Now new gives us a pointer, but we want the
		// value it points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(origValue.Type()).Elem()
		copyRecursive(copyValue, origValue)
		copy.Set(copyValue)

	case reflect.Struct:
		for i := 0; i < orig.NumField(); i += 1 {
			copyRecursive(copy.Field(i), orig.Field(i))
		}

	case reflect.Slice:
		copy.Set(reflect.MakeSlice(orig.Type(), orig.Len(), orig.Cap()))
		for i := 0; i < orig.Len(); i += 1 {
			copyRecursive(copy.Index(i), orig.Index(i))
		}

	case reflect.Map:
		copy.Set(reflect.MakeMap(orig.Type()))
		for _, key := range orig.MapKeys() {
			origValue := orig.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(origValue.Type()).Elem()
			copyRecursive(copyValue, origValue)
			copy.SetMapIndex(key, copyValue)
		}

	// And everything else will simply be taken from the original
	default:
		copy.Set(orig)
	}
}
