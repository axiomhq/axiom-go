// Code generated by "stringer -type=OwnerKind,QueryKind -linecomment -output=starred_string.go"; DO NOT EDIT.

package axiom

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[OwnedByUser-0]
	_ = x[OwnedByTeam-1]
}

const _OwnerKind_name = "OwnedByUserteam"

var _OwnerKind_index = [...]uint8{0, 11, 15}

func (i OwnerKind) String() string {
	if i >= OwnerKind(len(_OwnerKind_index)-1) {
		return "OwnerKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _OwnerKind_name[_OwnerKind_index[i]:_OwnerKind_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Analytics-1]
	_ = x[Stream-2]
}

const _QueryKind_name = "analyticsstream"

var _QueryKind_index = [...]uint8{0, 9, 15}

func (i QueryKind) String() string {
	i -= 1
	if i >= QueryKind(len(_QueryKind_index)-1) {
		return "QueryKind(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _QueryKind_name[_QueryKind_index[i]:_QueryKind_index[i+1]]
}
