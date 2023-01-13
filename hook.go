package exql

import (
	"context"
)

type HookList struct {
	list []Hook
}

func (h *HookList) Hook(ctx context.Context, query string, args ...any) {
	for _, hook := range h.list {
		hook.Hook(ctx, query, args...)
	}
}

func (h *HookList) Add(hook Hook) {
	h.list = append(h.list, hook)
}

func (h *HookList) Remove(hook Hook) {
	for i, v := range h.list {
		if v == hook {
			h.list = append(h.list[:i], h.list[i+1:]...)
			return
		}
	}
}
