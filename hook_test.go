package exql_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/mocks/mock_exql"
)

func TestHookList(t *testing.T) {
	h := &exql.HookList{}
	ctrl := gomock.NewController(t)
	hookA := mock_exql.NewMockHook(ctrl)
	hookA.EXPECT().Hook(gomock.Any(), "select", 1)
	h.Add(hookA)
	h.Hook(context.Background(), "select", 1)
	h.Remove(hookA)
	h.Hook(context.Background(), "select", 1)
}
