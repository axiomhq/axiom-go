// Code generated by "stringer -type=MessageCode,MessagePriority -linecomment -output=result_string.go"; DO NOT EDIT.

package query

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[emptyMessageCode-0]
	_ = x[VirtualFieldFinalizeError-1]
	_ = x[MissingColumn-2]
	_ = x[LicenseLimitForQueryWarning-3]
	_ = x[DefaultLimitWarning-4]
}

const _MessageCode_name = "virtual_field_finalize_errormissing_columnlicense_limit_for_query_warningdefault_limit_warning"

var _MessageCode_index = [...]uint8{0, 0, 28, 42, 73, 94}

func (i MessageCode) String() string {
	if i >= MessageCode(len(_MessageCode_index)-1) {
		return "MessageCode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MessageCode_name[_MessageCode_index[i]:_MessageCode_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[emptyMessagePriority-0]
	_ = x[Trace-1]
	_ = x[Debug-2]
	_ = x[Info-3]
	_ = x[Warn-4]
	_ = x[Error-5]
	_ = x[Fatal-6]
}

const _MessagePriority_name = "tracedebuginfowarnerrorfatal"

var _MessagePriority_index = [...]uint8{0, 0, 5, 10, 14, 18, 23, 28}

func (i MessagePriority) String() string {
	if i >= MessagePriority(len(_MessagePriority_index)-1) {
		return "MessagePriority(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MessagePriority_name[_MessagePriority_index[i]:_MessagePriority_index[i+1]]
}
