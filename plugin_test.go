package dokku

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type pluginManagerTestSuite struct {
	dokkuTestSuite
}

func TestRunPluginManagerTestSuite(t *testing.T) {
	suite.Run(t, new(pluginManagerTestSuite))
}

func (s *pluginManagerTestSuite) TestListPlugins() {
	r := s.Suite.Require()

	plugins, err := s.Client.ListPlugins()
	r.NoError(err)
	r.NotEmpty(plugins)
	r.NotEmpty(plugins[0])
}

func (s *pluginManagerTestSuite) TestInstallPlugin() {
	r := s.Suite.Require()

	// pluginName := "test-plugin"
	pluginURL := "https://github.com/dokku/dokku-mysql.git"
	err := s.Client.InstallPlugin(PluginInstallOptions{url: pluginURL})
	r.NoError(err)
}
