package resque_test

import (
	"github.com/steenzout/go-resque"
	"github.com/stretchr/testify/suite"
)

type ConsumerTestSuite struct {
	suite.Suite
}

// TestImplementsInterface tests if resque.BaseConsumer implements the resque.Consumer interface.
func (s *ConsumerTestSuite) TestImplementsInterface() {
	var _ resque.Consumer = (*resque.BaseConsumer)(nil)
}
