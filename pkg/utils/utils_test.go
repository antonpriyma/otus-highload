package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMin(t *testing.T) {
	type args struct {
		i1 int
		i2 int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "left is min",
			args: args{
				i1: 1,
				i2: 2,
			},
			want: 1,
		}, {
			name: "right is min",
			args: args{
				i1: 2,
				i2: 1,
			},
			want: 1,
		}, {
			name: "equal",
			args: args{
				i1: 1,
				i2: 1,
			},
			want: 1,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(tt *testing.T) {
			got := Min(testCase.args.i1, testCase.args.i2)
			require.Equal(tt, testCase.want, got)
		})
	}
}

func TestMax(t *testing.T) {
	type args struct {
		i1 int
		i2 int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "left is max",
			args: args{
				i1: 2,
				i2: 1,
			},
			want: 2,
		}, {
			name: "right is max",
			args: args{
				i1: 1,
				i2: 2,
			},
			want: 2,
		}, {
			name: "equal",
			args: args{
				i1: 1,
				i2: 1,
			},
			want: 1,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(tt *testing.T) {
			got := Max(testCase.args.i1, testCase.args.i2)
			require.Equal(tt, testCase.want, got)
		})
	}
}

func TestMinTime(tt *testing.T) {
	cases := []struct {
		t1, t2, expected time.Time
	}{
		{
			t1:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			t1:       MustParseTimeNano("2999-12-12T23:59:59.999999999+06:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			t1:       MustParseTimeNano("2999-12-12T23:59:59.999999998+07:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999998+07:00"),
		},
		{
			t1:       MustParseTimeNano("2999-12-12T23:59:58.999999999+07:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2999-12-12T23:59:58.999999999+07:00"),
		},
		{
			t1:       MustParseTimeNano("2999-12-12T23:58:59.999999999+07:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2999-12-12T23:58:59.999999999+07:00"),
		},
		{
			t1:       MustParseTimeNano("2999-12-12T22:59:59.999999999+07:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2999-12-12T22:59:59.999999999+07:00"),
		},
		{
			t1:       MustParseTimeNano("2999-12-11T23:59:59.999999999+07:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2999-12-11T23:59:59.999999999+07:00"),
		},
		{
			t1:       MustParseTimeNano("2999-11-12T23:59:59.999999999+07:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2999-11-12T23:59:59.999999999+07:00"),
		},
		{
			t1:       MustParseTimeNano("2998-12-12T23:59:59.999999999+07:00"),
			t2:       MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			expected: MustParseTimeNano("2998-12-12T23:59:59.999999999+07:00"),
		},
	}

	for _, c := range cases {
		c := c
		tt.Run(fmt.Sprintf("case_%s_%s", c.t1, c.t2), func(t *testing.T) {
			res := MinTime(c.t1, c.t2)
			require.Equal(t, c.expected, res)
			// inverse check
			res = MinTime(c.t2, c.t1)
			require.Equal(t, c.expected, res)
		})
	}
}
func TestMaxTime(tt *testing.T) {
	cases := []struct {
		tn       []time.Time
		expected time.Time
	}{
		{
			tn: []time.Time{
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			tn: []time.Time{
				MustParseTimeNano("2999-12-12T23:59:59.999999999+05:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+06:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+05:00"),
		},
		{
			tn: []time.Time{
				MustParseTimeNano("2999-12-12T23:59:59.999999997+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999998+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			tn: []time.Time{
				MustParseTimeNano("2999-12-12T23:59:57.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:58.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			tn: []time.Time{
				MustParseTimeNano("2999-12-12T23:57:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:58:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			tn: []time.Time{
				MustParseTimeNano("2999-12-12T21:59:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T22:59:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			tn: []time.Time{
				MustParseTimeNano("2999-12-10T23:59:59.999999999+07:00"),
				MustParseTimeNano("2999-12-11T23:59:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			tn: []time.Time{
				MustParseTimeNano("2999-10-12T23:59:59.999999999+07:00"),
				MustParseTimeNano("2999-11-12T23:59:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
		{
			tn: []time.Time{
				MustParseTimeNano("2997-12-12T23:59:59.999999999+07:00"),
				MustParseTimeNano("2998-12-12T23:59:59.999999999+07:00"),
				MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
			},
			expected: MustParseTimeNano("2999-12-12T23:59:59.999999999+07:00"),
		},
	}

	for _, c := range cases {
		c := c
		tt.Run(fmt.Sprintf("case_%s", c.expected), func(t *testing.T) {
			res := MaxTime(c.tn...)
			require.Equal(t, c.expected, res)
		})
	}
}

func TestMergeStringSlices(tt *testing.T) {
	cases := []struct {
		name                 string
		arr1, arr2, expected []string
	}{
		{
			name:     "merge with duppl",
			arr1:     []string{"a", "b"},
			arr2:     []string{"a", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "arr1 empty",
			arr1:     []string{},
			arr2:     []string{"a", "c"},
			expected: []string{"a", "c"},
		},
		{
			name:     "arr2 empty",
			arr1:     []string{"a", "c"},
			arr2:     []string{},
			expected: []string{"a", "c"},
		},
		{
			name:     "common merge",
			arr1:     []string{"a"},
			arr2:     []string{"b"},
			expected: []string{"a", "b"},
		},
	}

	for _, c := range cases {
		c := c
		tt.Run(c.name, func(t *testing.T) {
			res := MergeStringSlices(c.arr1, c.arr2)
			require.Equal(t, c.expected, res)
		})
	}
}
