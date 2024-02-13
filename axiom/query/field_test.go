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
	err := json.Unmarshal([]byte(`{ "type": "int|real" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, TypeInt|TypeReal, act.Type)
}

func TestFieldType_String(t *testing.T) {
	assert.Equal(t, TypeInvalid, FieldType(0))

	typ := TypeInt
	assert.Equal(t, "int", typ.String())

	typ |= TypeReal
	assert.Equal(t, "int|real", typ.String())
}

func TestFieldTypeFromString(t *testing.T) {
	for typ := TypeBool; typ <= TypeDictionary; typ <<= 1 {
		s := typ.String()

		parsedOp, err := fieldTypeFromString(s)
		if assert.NoError(t, err) {
			assert.NotEmpty(t, s)
			assert.Equal(t, typ, parsedOp)
		}
	}

	typ, err := fieldTypeFromString("abc")
	assert.Equal(t, TypeInvalid, typ)
	assert.EqualError(t, err, "unknown field type: abc")
}
