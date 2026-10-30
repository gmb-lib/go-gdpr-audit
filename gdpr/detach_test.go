package gdpr

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/gmb-lib/go-platform-kit/broker"
)

// Every string an envelope can carry is filled from a buffer of its own, every
// buffer is then reused, and every string of the copy must still read as it
// was. The fields are found by reflection, so a field the envelope gains later
// fails here until detach copies it.
func TestDetachSharesNoStringWithTheRecord(t *testing.T) {
	var bufs [][]byte

	next := func() string {
		buf := []byte(fmt.Sprintf("req-%02d", len(bufs)))
		bufs = append(bufs, buf)

		return unsafe.String(&buf[0], len(buf)) //nolint:gosec // the aliasing a request read makes, on purpose
	}

	var rec broker.Envelope
	fill(t, reflect.ValueOf(&rec).Elem(), next)

	out := detach(&rec)

	for _, buf := range bufs {
		copy(buf, strings.Repeat("x", len(buf)))
	}

	if n := check(t, reflect.ValueOf(out).Elem(), "Envelope"); n != len(bufs) {
		t.Errorf("read %d strings back, filled %d", n, len(bufs))
	}
}

func fill(t *testing.T, v reflect.Value, next func() string) {
	t.Helper()

	switch v.Kind() {
	case reflect.String:
		v.SetString(next())
	case reflect.Struct:
		if v.Type() == reflect.TypeFor[time.Time]() {
			return
		}

		for i := range v.NumField() {
			fill(t, v.Field(i), next)
		}
	case reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
		fill(t, v.Elem(), next)
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		fill(t, v.Index(0), next)
	case reflect.Map:
		if v.Type() != reflect.TypeFor[map[string]any]() {
			t.Fatalf("an envelope field of type %s: teach detach and this test to copy it", v.Type())
		}

		// The two kinds of attribute value a handler sets from a request.
		v.Set(reflect.ValueOf(map[string]any{
			next(): next(),
			next(): []string{next()},
		}))
	}
}

// check fails on every string under v that no longer reads as filled, and
// returns how many strings it read.
func check(t *testing.T, v reflect.Value, path string) int {
	t.Helper()

	switch v.Kind() {
	case reflect.String:
		if !strings.HasPrefix(v.String(), "req-") {
			t.Errorf("%s reads %q after the request's buffer was reused", path, v.String())
		}

		return 1
	case reflect.Struct:
		if v.Type() == reflect.TypeFor[time.Time]() {
			return 0
		}

		n := 0
		for i := range v.NumField() {
			n += check(t, v.Field(i), path+"."+v.Type().Field(i).Name)
		}

		return n
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			t.Errorf("%s is nil in the copy", path)

			return 0
		}

		return check(t, v.Elem(), path)
	case reflect.Slice:
		n := 0
		for i := range v.Len() {
			n += check(t, v.Index(i), fmt.Sprintf("%s[%d]", path, i))
		}

		return n
	case reflect.Map:
		n := 0
		for _, k := range v.MapKeys() {
			n += check(t, k, path+" key")
			n += check(t, v.MapIndex(k), fmt.Sprintf("%s[%s]", path, k))
		}

		return n
	}

	return 0
}
