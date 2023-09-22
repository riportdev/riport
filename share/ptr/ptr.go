package ptr

import (
	"time"

	"github.com/riportdev/riport/share/types"
)

func Time(t time.Time) *time.Time {
	return &t
}

func Bool(b bool) *bool {
	return &b
}

func String(s string) *string {
	return &s
}

func Int(i int) *int {
	return &i
}

func StringSlice(s ...string) *types.StringSlice {
	val := types.StringSlice(s)
	return &val
}
