package service

import "fmt"

var (
	Release string
	Dist    string
)

type Version struct {
	Release string
	Dist    string
}

func (v Version) String() string {
	return fmt.Sprintf("%s-%s", v.Release, v.Dist)
}

func (v Version) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", v)), nil
}
