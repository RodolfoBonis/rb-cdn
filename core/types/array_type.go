package types

type Array []any

func (a *Array) Contains(value any) bool {
	list := *a
	for i := 0; i < len(list); i++ {
		if list[i] == value {
			return true
		}
	}
	return false
}
