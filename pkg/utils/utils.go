package utils

import (
	"crypto/md5" //nolint: gosec
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"runtime"
	"sort"
	"strings"
	"time"
)

func Min(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}

func MinFloat(i1, i2 float64) float64 {
	if i1 < i2 {
		return i1
	}
	return i2
}

func Max(i1, i2 int) int {
	if i1 > i2 {
		return i1
	}
	return i2
}

func MaxFloat(i1, i2 float64) float64 {
	if i1 > i2 {
		return i1
	}
	return i2
}

func MD5SignatureWithSecret(queryParams url.Values, secret string) string {
	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var res strings.Builder
	for _, k := range keys {
		// Impossible error. String builder always returning nil as error
		MustFprintf(&res, "%s=%s", k, queryParams.Get(k))
	}

	sum := md5.Sum([]byte(res.String() + secret))
	return hex.EncodeToString(sum[:])
}

func MinTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}

func MaxTime(tn ...time.Time) time.Time {
	var maxTime time.Time
	for _, t := range tn {
		if t.After(maxTime) {
			maxTime = t
		}
	}
	return maxTime
}

func StringSliceToSet(slice []string) map[string]bool {
	ret := make(map[string]bool, len(slice))

	for _, v := range slice {
		ret[v] = true
	}

	return ret
}

func MD5Hex(toHash []byte) string {
	hashed := md5.Sum(toHash)
	return hex.EncodeToString(hashed[:])
}

func ReadString(r io.Reader) string {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Sprintf("failed to read from reader: %s", err)
	}

	return string(data)
}

func ContainsString(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}

	return false
}

func MergeStringSlices(arr1 []string, arr2 []string) []string {
	res := make([]string, 0, len(arr1))
	res = append(res, arr1...)

	setOfArr1 := StringSliceToSet(arr1)
	for _, str := range arr2 {
		if !setOfArr1[str] {
			res = append(res, str)
		}
	}

	return res
}

func Caller(skip int) string {
	pc, _, _, _ := runtime.Caller(skip + 1)
	funcName := runtime.FuncForPC(pc).Name()
	lastDot := strings.LastIndexByte(funcName, '.')
	return funcName[lastDot+1:]
}
