package expanders

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

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

// PointerWithDiags is Pointer for converters that also return diagnostics, so a call such as
// `field, diags = expanders.PointerWithDiags(convertXxxToAPI(ctx, list))` can wrap the
// converter's two results directly.
func PointerWithDiags[T any](v T, diags diag.Diagnostics) (*T, diag.Diagnostics) {
	return Pointer(v), diags
}
