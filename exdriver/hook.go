//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package exdriver

import (
	"context"
	"database/sql/driver"
)

// QueryHook is an effect-free middleware for queries
type QueryHook interface {
	HookQuery(ctx context.Context, query string, args []driver.NamedValue)
}

type HookList struct {
	list []QueryHook
}

func (h *HookList) HookQuery(ctx context.Context, query string, args []driver.NamedValue) {
	for _, hook := range h.list {
		hook.HookQuery(ctx, query, args)
	}
}

func (h *HookList) Add(hook QueryHook) {
	h.list = append(h.list, hook)
}

func (h *HookList) Remove(hook QueryHook) {
	for i, v := range h.list {
		if v == hook {
			h.list = append(h.list[:i], h.list[i+1:]...)
			return
		}
	}
}
