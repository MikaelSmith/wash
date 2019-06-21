package munge

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// ToType reinterprets the provided value v into the type pointer t. Returns an error if the type
// cannot be converted. This exists primarily to get around JSON Unmarshal's tendency to create
// []interface{} and map[string]interface{} rather than use something more specific.
//
// All fields in v must be mappable to fields in t.
func ToType(v interface{}, t interface{}) error {
	// Check that we can unmarshal into t.
	rt := reflect.ValueOf(t)
	if rt.Kind() != reflect.Ptr || rt.IsNil() {
		return &json.InvalidUnmarshalError{Type: rt.Type()}
	}

	newVal, err := toType(reflect.ValueOf(v), rt.Elem().Type())
	if err != nil {
		return err
	}
	rt.Elem().Set(newVal)
	return nil
}

func toType(v reflect.Value, dt reflect.Type) (reflect.Value, error) {
	newV := reflect.ValueOf(v.Interface())
	if newV.Type().ConvertibleTo(dt) {
		return newV.Convert(dt), nil
	}

	if v.Kind() == reflect.Slice && dt.Kind() == reflect.Slice {
		destSlice := reflect.MakeSlice(dt, v.Len(), v.Len())
		for i := 0; i < destSlice.Len(); i++ {
			newEntry, err := toType(v.Index(i), dt.Elem())
			if err != nil {
				return destSlice, err
			}
			destSlice.Index(i).Set(newEntry)
		}
		return destSlice, nil
	}

	if v.Kind() == reflect.Map && dt.Kind() == reflect.Map {
		destMap := reflect.MakeMapWithSize(dt, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			newEntry, err := toType(iter.Value(), dt.Elem())
			if err != nil {
				return destMap, err
			}
			destMap.SetMapIndex(iter.Key(), newEntry)
		}
		return destMap, nil
	}

	return newV, fmt.Errorf("cannot convert %v to %v", v.Type(), dt)
}
