package test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func CheckError(t *testing.T, err, exactError error, wantAnyError bool) bool {
	switch {
	case exactError != nil:
		err = ExtractExactError(err, exactError)

		// errors in cause fields have different stacks (by WithStack)
		// they will never be equal, even they was created by the same way
		// set them to nil
		exactError = setFieldIfPossible(exactError, "Cause", new(error)).(error)
		err = setFieldIfPossible(err, "Cause", new(error)).(error)

		require.Equal(t, exactError, err)
		return true
	case wantAnyError:
		require.Error(t, err)
		return true
	default:
		require.NoError(t, err)
		return false
	}
}

func ExtractExactError(err, exactErr error) error {
	if errors.Is(err, exactErr) {
		return exactErr
	}

	errType := reflect.TypeOf(exactErr)
	errValue := reflect.New(errType) // New() returns pointer of err type
	if !errors.As(err, errValue.Interface()) {
		return err
	}

	return errValue.Elem().Interface().(error)
}

func setFieldIfPossible(obj interface{}, field string, newValue interface{}) interface{} {
	value := reflect.ValueOf(obj)
	retValue := value

	if value.Kind() == reflect.Struct {
		// this is equal to copy with getting pointer
		// value := &errorType{errorValue}
		ptrValue := reflect.New(value.Type())
		ptrValue.Elem().Set(value)

		value = ptrValue
		retValue = value.Elem()
	}

	if value.Kind() != reflect.Ptr {
		return obj
	}
	value = value.Elem()

	if value.Kind() != reflect.Struct {
		return obj
	}

	f := value.FieldByName(field)
	if !f.IsValid() || f.IsZero() {
		// field not found
		return obj
	}
	if !f.CanSet() {
		// field is not setable
		return obj
	}
	if !f.Type().AssignableTo(reflect.TypeOf(newValue).Elem()) {
		// field and value are of different types
		return obj
	}

	// set field to value
	f.Set(reflect.ValueOf(newValue).Elem())
	return retValue.Interface()
}

func RequireEqualAsJSON(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	require.Equal(t, structToJSONMap(t, expected), structToJSONMap(t, actual), msgAndArgs...)
}

func structToJSONMap(t *testing.T, in interface{}) interface{} {
	encoded, err := json.Marshal(in)
	require.NoError(t, err)

	var decoded interface{}
	if reflect.ValueOf(in).Kind() == reflect.Slice {
		decoded = new([]map[string]interface{})
	} else {
		decoded = new(map[string]interface{})
	}

	err = json.Unmarshal(encoded, decoded)
	require.NoError(t, err)

	// dereference a pointer
	return reflect.ValueOf(decoded).Elem().Interface()
}
