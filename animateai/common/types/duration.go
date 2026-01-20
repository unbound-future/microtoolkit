package types

import (
	"fmt"
	"strings"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`) // 去掉可能的引号
	if s == "" || s == "null" {
		d.Duration = 0
		return nil
	}
	// 尝试直接解析成 duration，比如 "5s" 或 5s
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = dur
	return nil
}