package resque_test

import (
	"github.com/steenzout/go-resque"
	"github.com/stretchr/testify/suite"
)

type QueueTestSuite struct {
	suite.Suite
}

// TestImplementsInterface tests if resque.RedisQueue implements the resque.Queue interface.
func (s *QueueTestSuite) TestImplementsInterface() {
	var _ resque.Queue = (*resque.RedisQueue)(nil)
}
