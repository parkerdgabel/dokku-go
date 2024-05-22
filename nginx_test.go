package dokku

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type nginxManagerTestSuite struct {
	dokkuTestSuite
}

func TestRunNginxManagerTestSuite(t *testing.T) {
	suite.Run(t, new(networkManagerTestSuite))
}

func (s *nginxManagerTestSuite) TestGetAppConfig() {
	r := s.Suite.Require()

	testApp := "test-nginx-app"
	r.NoError(s.Client.CreateApp(testApp))

	_, err := s.Client.GetAppNginxConfig(testApp)
	r.ErrorIs(err, NginxNoConfigErr)
}
