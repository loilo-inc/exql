package exql

import (
	"fmt"
	"strings"
)

func ParseTags(tag string) (map[string]string, error) {
	tags := strings.Split(tag, ";")
	ret := make(map[string]string)
	set := func(k string, v string) error {
		if k == "" {
			return nil
		}
		if _, ok := ret[k]; ok {
			return fmt.Errorf("duplicate tag: %s", k)
		}
		ret[k] = v
		return nil
	}
	for _, tag := range tags {
		kv := strings.Split(tag, ":")
		if len(kv) == 0 {
			continue
		}
		if len(kv) == 1 {
			if err := set(kv[0], ""); err != nil {
				return nil, err
			}
		} else if len(kv) == 2 {
			if err := set(kv[0], kv[1]); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("invalid tag format: %s", tag)
		}
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("empty tags")
	}
	return ret, nil
}