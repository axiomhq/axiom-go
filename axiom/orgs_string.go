// Code generated by "stringer -type=PaymentStatus -linecomment -output=orgs_string.go"; DO NOT EDIT.

package axiom

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[emptyPaymentStatus-0]
	_ = x[Success-1]
	_ = x[NotAvailable-2]
	_ = x[Failed-3]
	_ = x[Blocked-4]
}

const _PaymentStatus_name = "successnafailedblocked"

var _PaymentStatus_index = [...]uint8{0, 0, 7, 9, 15, 22}

func (i PaymentStatus) String() string {
	if i >= PaymentStatus(len(_PaymentStatus_index)-1) {
		return "PaymentStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _PaymentStatus_name[_PaymentStatus_index[i]:_PaymentStatus_index[i+1]]
}
