package stat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCallerLabel(t *testing.T) {
	func() { //WithCaller
		require.Equal(t, Labels{"method": "TestCallerLabel"}, CallerLabel())
	}()
}
