package resque_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// PackageTestSuite test suite for the resque package.
type PackageTestSuite struct {
	suite.Suite
}

// Test runs package tests.
func Test(t *testing.T) {
	// ordered by dependencies
	suite.Run(t, new(ConsumerTestSuite))
	suite.Run(t, new(QueueTestSuite))
	suite.Run(t, new(ProducerTestSuite))
}
