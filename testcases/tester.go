package testcases

import (
	"context"

	"github.com/iftechio/vk-test/suite"
)

type Tester interface {
	Name() string
	Test(ctx context.Context, s *suite.Suite) error
}
