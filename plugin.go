package dokku

import (
	"fmt"
	"regexp"
	"strings"
)

type pluginManager interface {
	/*
		EnablePlugin(plugin string) error
		DisablePlugin(plugin string) error

		CheckPluginInstalled(plugin string) (bool, error)
		InstallPlugin(options PluginInstallOptions) error
		InstallPluginDependencies() error
		UninstallPlugin(plugin string) error
		UpdatePlugin(plugin string) error
		UpdatePlugins() error

		TriggerPluginHook(hookArgs []string) error
	*/
	InstallPlugin(options PluginInstallOptions) error
	ListPlugins() ([]PluginInfo, error)
}

type PluginInfo struct {
	Name        string
	Version     string
	Enabled     bool
	Description string
}

const (
	pluginInstalledCmd           = "plugin:installed %s"
	pluginDisableCmd             = "plugin:disable %s"
	pluginEnableCmd              = "plugin:enable %s"
	pluginInstallFullCmd         = "plugin:install %s --committish %s --name %s"
	pluginInstallCmd             = "plugin:install %s"
	pluginInstallWithNameCmd     = "plugin:install %s --name %s"
	pluginInstallGitCmd          = "plugin:install %s --committish %s"
	pluginInstallDependenciesCmd = "plugin:install-dependencies"
	pluginListCmd                = "plugin:list"
	pluginTriggerCmd             = "plugin:trigger %s"
	pluginUninstallCmd           = "plugin:uninstall %s"
	pluginUpdateCmd              = "plugin:update %s %s"
)

func (c *BaseClient) ListPlugins() ([]PluginInfo, error) {
	out, err := c.Exec(pluginListCmd)
	lines := strings.Split(out, "\n")
	plugins := make([]PluginInfo, len(lines))
	var multipleWhitespaceRe = regexp.MustCompile("\\s+")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		cols := multipleWhitespaceRe.Split(line, 4)
		if len(cols) < 4 {
			return nil, fmt.Errorf("error parsing plugin list line: %s", line)
		}

		plugins[i] = PluginInfo{
			Name:        cols[0],
			Version:     cols[1],
			Enabled:     cols[2] == "enabled",
			Description: cols[3],
		}
	}

	return plugins, err
}

type PluginInstallOptions struct {
	url        string `dokku:"plugin-url"`
	committish string `dokku:"committish"`
	name       string `dokku:"plugin-name"`
}

func (c *BaseClient) InstallPlugin(options PluginInstallOptions) error {
	if options.url == "" {
		return fmt.Errorf("plugin url is required")
	}
	if strings.HasPrefix(options.url, "git@") && options.committish != "" {
		cmd := fmt.Sprintf(pluginInstallGitCmd, options.url, options.committish)
		_, err := c.Exec(cmd)
		return err
	} else if options.committish != "" && options.name != "" {
		cmd := fmt.Sprintf(pluginInstallFullCmd, options.url, options.committish, options.name)
		_, err := c.Exec(cmd)
		return err
	} else if options.name != "" {
		cmd := fmt.Sprintf(pluginInstallWithNameCmd, options.url, options.name)
		_, err := c.Exec(cmd)
		return err
	} else {
		cmd := fmt.Sprintf(pluginInstallCmd, options.url)
		_, err := c.Exec(cmd)
		return err
	}
}

/*
func (c *BaseClient) CheckPluginInstalled(plugin string) (bool, error) {
	cmd := fmt.Sprintf(pluginEnableCmd, plugin)
	out, err := c.Exec(cmd)
	fmt.Println(out)
	return false, err
}

func (c *BaseClient) EnablePlugin(plugin string) error {
	cmd := fmt.Sprintf(pluginEnableCmd, plugin)
	_, err := c.Exec(cmd)
	return err
}

func (c *BaseClient) DisablePlugin(plugin string) error {
	cmd := fmt.Sprintf(pluginDisableCmd, plugin)
	_, err := c.Exec(cmd)
	return err
}

func (c *BaseClient) InstallPlugin(options PluginInstallOptions) error {
	//TODO implement me
	panic("implement me")
}

func (c *BaseClient) InstallPluginDependencies() error {
	_, err := c.Exec(pluginInstallDependenciesCmd)
	return err
}

func (c *BaseClient) UninstallPlugin(plugin string) error {
	cmd := fmt.Sprintf(pluginUninstallCmd, plugin)
	_, err := c.Exec(cmd)
	return err
}

func (c *BaseClient) UpdatePlugin(plugin string) error {
	cmd := fmt.Sprintf(pluginUpdateCmd, plugin)
	_, err := c.Exec(cmd)
	return err
}

func (c *BaseClient) UpdatePlugins() error {
	//TODO implement me
	panic("implement me")
}

func (c *BaseClient) TriggerPluginHook(hookArgs []string) error {
	//TODO implement me
	panic("implement me")
}
*/
