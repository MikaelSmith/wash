package munge

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToTypeStringsToStrings(t *testing.T) {
	v := []string{"hello", "there"}
	var d []string
	assert.NoError(t, ToType(v, &d))
	assert.Equal(t, v, d)
}

func BenchmarkToTypeStringsToStrings(b *testing.B) {
	v := []string{"hello", "there"}
	var d []string
	for n := 0; n < b.N; n++ {
		ToType(v, &d)
	}
}

func BenchmarkConvertStringsToStrings(b *testing.B) {
	v := []string{"hello", "there"}
	vi := interface{}(v)
	var d []string
	for n := 0; n < b.N; n++ {
		if val, ok := vi.([]string); ok {
			*(&d) = val
		}
	}
}

func BenchmarkMarshalStringsToStrings(b *testing.B) {
	v := []string{"hello", "there"}
	var d []string
	for n := 0; n < b.N; n++ {
		b, _ := json.Marshal(v)
		json.Unmarshal(b, &d)
	}
}

func TestToTypeStringMapToStringMap(t *testing.T) {
	v := map[string]string{"hello": "there"}
	var d map[string]string
	assert.NoError(t, ToType(v, &d))
	assert.Equal(t, v, d)
}

func BenchmarkToTypeStringMapToStringMap(b *testing.B) {
	v := map[string]string{"hello": "there"}
	var d map[string]string
	for n := 0; n < b.N; n++ {
		ToType(v, &d)
	}
}

func BenchmarkConvertStringMapToStringMap(b *testing.B) {
	v := map[string]string{"hello": "there"}
	vi := interface{}(v)
	var d map[string]string
	for n := 0; n < b.N; n++ {
		if val, ok := vi.(map[string]string); ok {
			*(&d) = val
		}
	}
}

func BenchmarkMarshalStringMapToStringMap(b *testing.B) {
	v := map[string]string{"hello": "there"}
	var d map[string]string
	for n := 0; n < b.N; n++ {
		b, _ := json.Marshal(v)
		json.Unmarshal(b, &d)
	}
}

func TestToTypeInterfacesToStrings(t *testing.T) {
	v := []interface{}{"hello", "there"}
	var d []string
	assert.NoError(t, ToType(v, &d))
	assert.Equal(t, []string{"hello", "there"}, d)
}

func BenchmarkToTypeInterfacesToStrings(b *testing.B) {
	v := []interface{}{"hello", "there"}
	var t []string
	for n := 0; n < b.N; n++ {
		ToType(v, &t)
	}
}

func BenchmarkConvertInterfacesToStrings(b *testing.B) {
	v := []interface{}{"hello", "there"}
	vi := interface{}(v)
	var d []string
	for n := 0; n < b.N; n++ {
		if arr, ok := vi.([]interface{}); ok {
			newarr := make([]string, len(arr))
			for i, val := range arr {
				if entry, ok := val.(string); ok {
					newarr[i] = entry
				}
			}
			*(&d) = newarr
		}
	}
}

func BenchmarkMarshalInterfacesToStrings(b *testing.B) {
	v := []interface{}{"hello", "there"}
	var d []string
	for n := 0; n < b.N; n++ {
		b, _ := json.Marshal(v)
		json.Unmarshal(b, &d)
	}
}

func TestToTypeInterfaceMapToStringMap(t *testing.T) {
	v := map[string]interface{}{"hello": "there"}
	var d map[string]string
	assert.NoError(t, ToType(v, &d))
	assert.Equal(t, map[string]string{"hello": "there"}, d)
}

func BenchmarkToTypeInterfaceMapToStringMap(b *testing.B) {
	v := map[string]interface{}{"hello": "there"}
	var d map[string]string
	for n := 0; n < b.N; n++ {
		ToType(v, &d)
	}
}

func BenchmarkConvertInterfaceMapToStringMap(b *testing.B) {
	v := map[string]interface{}{"hello": "there"}
	vi := interface{}(v)
	var d map[string]string
	for n := 0; n < b.N; n++ {
		if mp, ok := vi.(map[string]interface{}); ok {
			newmp := make(map[string]string, len(mp))
			for k, val := range mp {
				if entry, ok := val.(string); ok {
					newmp[k] = entry
				}
			}
			*(&d) = newmp
		}
	}
}

func BenchmarkMarshalInterfaceMapToStringMap(b *testing.B) {
	v := map[string]interface{}{"hello": "there"}
	var d map[string]string
	for n := 0; n < b.N; n++ {
		b, _ := json.Marshal(v)
		json.Unmarshal(b, &d)
	}
}

func TestToTypeWithInconvertibleTypesError(t *testing.T) {

}
