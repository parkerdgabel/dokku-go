package dokku

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type gitManagerTestSuite struct {
	dokkuTestSuite
}

func TestRunGitManagerTestSuite(t *testing.T) {
	suite.Run(t, &gitManagerTestSuite{
		dokkuTestSuite{
			DefaultAppName:            "test-git-app",
			AttachContainerTestLogger: true,
		},
	})
}

func (s *gitManagerTestSuite) TestGitReport() {
	r := s.Suite.Require()
	var err error

	report, err := s.Client.GitGetAppReport(s.DefaultAppName)
	r.NoError(err)
	r.Equal("master", report.DeployBranch)
}

func (s *gitManagerTestSuite) TestSyncGitRepo() {
	r := s.Suite.Require()
	var err error

	//ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	r.NoError(s.Dokku.InstallBuildPacksCLI(context.Background()))

	r.NoError(s.Client.DisableAppDeployChecks(s.DefaultAppName))

	testRepo := "https://github.com/parkerdgabel/go-hello-world-http.git"
	options := &GitSyncOptions{
		Build:  true,
		GitRef: "main",
	}
	stream, err := s.Client.GitSyncAppRepo(s.DefaultAppName, testRepo, options)
	r.NoError(err)
	r.NotEmpty(stream.Stdout)
}
