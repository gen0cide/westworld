package runtime

// absInt returns the absolute value of v. Used for Chebyshev /
// Manhattan distance checks throughout the runtime layer.
//
// Previously declared in runtime/combat_loop.go; lifted here when
// that file was removed (the routine-language `select{}` +
// `when{}` constructs replaced the Go reactor — see
// examples/routines/combat_loop.routine).
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
