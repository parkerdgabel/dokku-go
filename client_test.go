package dokku

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type dokkuClientTestSuite struct {
	suite.Suite
}

func TestRunDokkuClientTestSuite(t *testing.T) {
	suite.Run(t, new(dokkuClientTestSuite))
}

func (s *checksManagerTestSuite) TestSSHClientExecStreaming() {
	r := s.Suite.Require()
	stream, err := s.Client.ExecStreaming("version")
	r.NoError(err)
	output, err := ioutil.ReadAll(stream.Stdout)
	r.NoError(err)
	r.NotEmpty(output)
}

func (s *checksManagerTestSuite) TestSSHClientExecStreamingError() {
	r := s.Suite.Require()
	stream, err := s.Client.ExecStreaming("bad command")
	r.NoError(err)
	r.NoError(stream.Error)
	time.Sleep(100 * time.Millisecond)
	r.Error(stream.Error)
}
