package utils

// delete set ig
type Set struct {
	set map[any]bool
}

func (Set) New() Set {
	return Set{}
}
