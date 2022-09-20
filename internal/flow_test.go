package internal

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testFlowSuite struct {
	suite.Suite
}

func (s *testFlowSuite) SetupAllSuite() {

}

func (s testFlowSuite) Test_genIssueName() {
	name := genIssueBranchName("milestone-test", 123)
	s.Equal(IssueBranchPrefix+"milestone-test-123", name)

	feature := parseFeatureFromIssueName(name, false)
	s.Equal("milestone-test", feature)

	feature2, ok := tryParseFeatureNameFrom(name, false)
	s.True(ok)
	s.Equal("milestone-test", feature2)
}

func (s testFlowSuite) Test_parseFeatureFromIssueName_compatible() {
	name := "123-milestone-test"
	feature := parseFeatureFromIssueName(name, true)
	s.Equal("milestone-test", feature)

}

func Test_flowSuite(t *testing.T) {
	suite.Run(t, new(testFlowSuite))
}
