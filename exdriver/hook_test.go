package exdriver_test

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2/exdriver"
	"github.com/loilo-inc/exql/v2/mocks/mock_exdriver"
)

func TestHookList(t *testing.T) {
	h := &exdriver.HookList{}
	ctrl := gomock.NewController(t)
	hookA := mock_exdriver.NewMockQueryHook(ctrl)
	args := []driver.NamedValue{{Ordinal: 1, Value: 1}}
	hookA.EXPECT().HookQuery(gomock.Any(), "select", args)
	h.Add(hookA)
	h.HookQuery(context.Background(), "select", args)
	h.Remove(hookA)
	h.HookQuery(context.Background(), "select", args)
}
