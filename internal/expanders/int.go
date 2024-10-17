package expanders

func IntPointer[I int32 | int64](i *I) *int {
	if i == nil {
		return nil
	}

	intValue := int(*i)
	return &intValue
}
