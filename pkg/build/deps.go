//go:build tools

package build

import (
	_ "github.com/golang/mock/mockgen"       // needed for mocks generation
	_ "github.com/golang/mock/mockgen/model" // needed for mocks generation
	_ "golang.org/x/tools/cmd/goimports"     // needed for import
)
