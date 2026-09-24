package expanders

import "reflect"

// Pointer returns a pointer to v for SDK request fields declared as *T, or nil when v is
// itself nil (a nil slice, map, pointer, interface, chan or func). Returning nil matters
// for `omitempty`: a nil *[]T leaves the field out of the request, while &v with a nil
// slice would be sent as an explicit null. Any non-nil value — including an empty slice,
// which is then sent as [] — becomes &v.
func Pointer[T any](v T) *T {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil // v is a nil interface
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Pointer, reflect.Interface, reflect.Chan, reflect.Func:
		if rv.IsNil() {
			return nil
		}
	}
	return &v
}
