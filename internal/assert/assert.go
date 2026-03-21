package assert

import (
	"reflect"
	"testing"
)

func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if !IsEqual(got, want) {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func IsEqual[T any](got, want T) bool {
	if isNil(got) && isNil(want) {
		return true
	}
	return reflect.DeepEqual(got, want)
}

func isNil(v any) bool {
	// Returns true if v equals nil.
	if v == nil {
		return true
	}
	// Use reflection to check the underlying type of v, and return true if it
	// is a nullable type (e.g. pointer, map or slice) with a value of nil.
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	}
	// Other types like string, bool, int are never nil.
	return false
}

func Nil(t *testing.T, got any) {
	t.Helper()
	if !isNil(got) {
		t.Errorf("got: %v; want: nil", got)
	}
}

func NotEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if IsEqual(got, want) {
		t.Errorf("got: %v; expected values to be different", got)
	}
}
func True(t *testing.T, got bool) {
	t.Helper()
	if !got {
		t.Errorf("got: false; want: true")
	}
}
func False(t *testing.T, got bool) {
	t.Helper()
	if got {
		t.Errorf("got: true; want: false")
	}
}

func NotNil(t *testing.T, got any) {
	t.Helper()
	if isNil(got) {
		t.Errorf("got: nil; want: non-nil")
	}
}
