package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldType_Unmarshal(t *testing.T) {
	var act struct {
		Type FieldType `json:"type"`
	}
	err := json.Unmarshal([]byte(`{ "type": "int|string" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, TypeInteger|TypeString, act.Type)
}

func TestFieldType_String(t *testing.T) {
	assert.Equal(t, TypeInvalid, FieldType(0))

	typ := TypeDateTime
	assert.Equal(t, "datetime", typ.String())

	typ |= TypeTimespan
	assert.Equal(t, "datetime|timespan", typ.String())
}

func TestFieldType_Bool(t *testing.T) {
	assert.Equal(t, TypeInvalid, FieldType(0))

	assert.Equal(t, "boolean", TypeBool.String())
}

func TestFieldTypeFromString(t *testing.T) {
	for typ := TypeBool; typ <= TypeUnknown; typ <<= 1 {
		s := typ.String()

		parsedOp, err := fieldTypeFromString(s)
		if assert.NoError(t, err) {
			assert.NotEmpty(t, s)
			assert.Equal(t, typ, parsedOp)
		}
	}

	typ, err := fieldTypeFromString("abc")
	assert.Equal(t, TypeInvalid, typ)
	assert.EqualError(t, err, "invalid field type: abc")
}
