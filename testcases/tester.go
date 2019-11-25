package testcases

import (
	"context"

	"github.com/iftechio/vk-test/suite"
)

type Tester interface {
	Name() string
	Description() string
	Test(ctx context.Context, s *suite.Suite) error
}
