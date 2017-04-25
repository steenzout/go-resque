package resque_test

import (
	"github.com/steenzout/go-resque"
	"github.com/stretchr/testify/suite"
)

type ProducerTestSuite struct {
	suite.Suite
}

// TestImplementsInterface tests if resque.BaseProducer implements the resque.Producer interface.
func (s *ProducerTestSuite) TestImplementsInterface() {
	var _ resque.Producer = (*resque.BaseProducer)(nil)
}
