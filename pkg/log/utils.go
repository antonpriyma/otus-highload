package log

import (
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

func AsJSON(v interface{}) fmt.Stringer {
	return asJSON{Value: v}
}

type asJSON struct {
	Value interface{} `json:"value"`
}

func (j asJSON) String() string {
	// json with struct, because value can be scalar
	return toJSON(j)
}

func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// json may only fail if custom marshaler defined or scalar value was given
		return fmt.Sprintf("json failed: %s, spew: %s", err, toSPEW(v))
	}

	return string(data)
}

var spewParametrized = spew.ConfigState{
	Indent:                  "",
	MaxDepth:                100,
	DisableMethods:          true,
	DisablePointerMethods:   true,
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	ContinueOnMethod:        false,
	SortKeys:                false,
}

func toSPEW(v interface{}) string {
	return spewParametrized.Sprintf("%#v", v)
}
