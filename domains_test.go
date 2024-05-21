package dokku

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type domainsManagerTestSuite struct {
	dokkuTestSuite
}

func (s *domainsManagerTestSuite) AfterTest(suiteName, testName string) {
	s.Client.DestroyApp("test-domains-app")
	s.Client.ClearAppDomains("test-domains-app")

}
func TestRunDomainsManagerTestSuite(t *testing.T) {
	suite.Run(t, new(domainsManagerTestSuite))
}

func (s *domainsManagerTestSuite) TestGetAppDomains() {
	r := s.Require()

	testAppName := "test-domains-app"
	r.NoError(s.Client.CreateApp(testAppName))

	appDomain := "foo.example.com"
	globalDomain := "bar.example.com"

	r.NoError(s.Client.AddAppDomain(testAppName, appDomain))
	r.NoError(s.Client.AddGlobalDomain(globalDomain))

	report, err := s.Client.GetAppDomainsReport(testAppName)
	r.NoError(err)

	r.Len(report.AppDomains, 1)
	r.Equal(report.AppDomains[0], appDomain)

	r.Len(report.GlobalDomains, 1)
	r.Equal(report.GlobalDomains[0], globalDomain)
}

func (s *domainsManagerTestSuite) TestListNoAppDomains() {
	r := s.Require()

	testAppName := "test-domains-app"
	r.NoError(s.Client.CreateApp(testAppName))

	r.NoError(s.Client.DisableAppDomains(testAppName))

	report, err := s.Client.GetAppDomainsReport(testAppName)
	r.NoError(err)
	r.Len(report.AppDomains, 0)
}
