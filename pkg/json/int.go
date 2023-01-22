package json

import (
	"encoding/json"
	"strconv"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

type Int int

func (i *Int) UnmarshalJSON(bytes []byte) error {
	if bytes[0] != '"' {
		return json.Unmarshal(bytes, (*int)(i))
	}

	var strInt string
	if err := json.Unmarshal(bytes, &strInt); err != nil {
		return errors.Wrap(err, "failed to unmarshal int as string")
	}

	parsedInt, err := strconv.Atoi(strInt)
	if err != nil {
		return errors.Wrapf(err, "failed to convert string to int: %v", strInt)
	}

	*i = Int(parsedInt)

	return nil
}

type IntSlice []Int

func (is IntSlice) GoIntSlice() []int {
	ret := make([]int, 0, len(is))
	for _, v := range is {
		ret = append(ret, int(v))
	}

	return ret
}

func IntSliceFromGo(s []int) IntSlice {
	ret := make(IntSlice, 0, len(s))
	for _, v := range s {
		ret = append(ret, Int(v))
	}

	return ret
}
