package testutils

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testingImage     = "dokku/dokku:latest"
	dockerSocketFile = "/var/run/docker.sock"
)

type nullLogger struct{}

func (l nullLogger) Printf(format string, args ...any) {}

func CreateDokkuContainer(ctx context.Context, withLogs bool) (*DokkuContainer, error) {
	if runtime.GOOS == "darwin" {
		if err := setupDarwinEnv(); err != nil {
			return nil, err
		}
	}

	mounts := testcontainers.ContainerMounts{
		// mounting the docker socket into the container is insecure, but nobody else should run this
		testcontainers.BindMount(dockerSocketFile, dockerSocketFile),
		testcontainers.VolumeMount("dokku-ssl", "/mnt/dokku/etc/nginx"),
	}

	req := testcontainers.ContainerRequest{
		Image: testingImage,
		// FromDockerfile: testcontainers.FromDockerfile{Context: "./internal/testutils", Dockerfile: "Dockerfile"},
		ExposedPorts: []string{"22/tcp", "80/tcp", "443/tcp"},
		Mounts:       mounts,
		WaitingFor:   wait.ForListeningPort("22").WithStartupTimeout(30 * time.Second),
	}

	var logger testcontainers.Logging
	if !withLogs {
		logger = nullLogger{}
	}

	gReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Logger:           logger,
	}

	container, err := testcontainers.GenericContainer(ctx, gReq)
	if err != nil {
		return nil, err
	}

	if err := ensureMatchingDockerGroupId(ctx, container); err != nil {
		return nil, maybeTerminateContainerAfterError(ctx, container, err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, maybeTerminateContainerAfterError(ctx, container, err)
	}

	mappedSSHPort, err := container.MappedPort(ctx, "22")
	if err != nil {
		return nil, maybeTerminateContainerAfterError(ctx, container, err)
	}

	rootKeyPair, err := GenerateRSAKeyPair()
	if err != nil {
		return nil, maybeTerminateContainerAfterError(ctx, container, err)
	}

	dc := &DokkuContainer{
		Container:      container,
		Host:           host,
		SSHPort:        mappedSSHPort.Port(),
		RootPublicKey:  rootKeyPair.PublicKey,
		RootPrivateKey: rootKeyPair.PrivateKey,
	}
	dc.RegisterRootAuthorizedKey(ctx)

	// dc.ConfigureSSHD(ctx)
	// reader, err := dc.CopyFileFromContainer(ctx, "/etc/ssh/sshd_config")
	// if err != nil {
	// 	return nil, maybeTerminateContainerAfterError(ctx, container, err)
	// }
	// defer reader.Close()
	// scanner := bufio.NewScanner(reader)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if strings.Contains(line, "PermitRootLogin") {
	// 		fmt.Println(line)
	// 	}
	// }
	// dc.RestartSSHD(ctx)

	return dc, nil
}

func setupDarwinEnv() error {

	localDockerSocketFile := "/var/run/docker.sock"
	localDockerSocketURI := fmt.Sprintf("unix://%s", localDockerSocketFile)

	if err := os.Setenv("DOCKER_HOST", localDockerSocketURI); err != nil {
		return err
	}
	if err := os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", localDockerSocketFile); err != nil {
		return err
	}

	return nil
}

func ensureMatchingDockerGroupId(ctx context.Context, container testcontainers.Container) error {
	exitCode, err := container.Exec(ctx, []string{"groupmod", "-g", "99", "systemd-timesync"})
	if exitCode != 0 {
		return fmt.Errorf("failed to change gid of containerized systemd-timesync group, got exit code %d\n", exitCode)
	} else if err != nil {
		return err
	}

	exitCode, err = container.Exec(ctx, []string{"groupmod", "-g", "101", "docker"})
	if exitCode != 0 {
		return fmt.Errorf("failed to change gid of containerized docker group, got exit code %d\n", exitCode)
	}

	return err
}

func maybeTerminateContainerAfterError(ctx context.Context, container testcontainers.Container, err error) error {
	if termErr := container.Terminate(ctx); termErr != nil {
		return fmt.Errorf("failed to terminate container: %s after failing to handle error: %w", termErr.Error(), err)
	}
	return err
}
