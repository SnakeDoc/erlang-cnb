package erlang_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitErlang(t *testing.T) {
	suite := spec.New("erlang", spec.Report(report.Terminal{}))
	suite("ErlangVersionResolver", testErlangVersionResolver)
	suite.Run(t)
}
