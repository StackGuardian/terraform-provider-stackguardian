package flatteners

// PointerValue returns *p, or the zero value of T when p is nil. For an SDK *[]T field
// the zero value is a nil slice, so a field absent from the response stays nil (read back
// as null) while a pointer to an empty slice stays [] — converters such as
// ListOfStringToTerraformList can still tell the two apart.
func PointerValue[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
