package testutils

import (
	"bytes"
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

	rootKeyPair, err := GenerateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate root key pair: %w", err)
	}

	req := testcontainers.ContainerRequest{
		Image: testingImage,
		// FromDockerfile: testcontainers.FromDockerfile{Context: "./internal/testutils", Dockerfile: "Dockerfile"},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            bytes.NewReader(rootKeyPair.PublicKey),
				ContainerFilePath: "/root/.ssh/authorized_keys",
				FileMode:          0600,
			},
		},
		ExposedPorts: []string{"22/tcp", "80/tcp", "443/tcp"},
		Mounts:       mounts,
		WaitingFor:   wait.ForListeningPort("22").WithStartupTimeout(30 * time.Second),
	}
	// privateKeyBytes := encodePrivateKeyToPEM(rootKeyPair.PrivateKey)

	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal private key: %w", err)
	// }
	// reader := bytes.NewReader(privateKeyBytes)
	// scanner := bufio.NewScanner(reader)
	// for scanner.Scan() {
	// 	fmt.Println(scanner.Text())
	// }

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

	dc := &DokkuContainer{
		Container:      container,
		Host:           host,
		SSHPort:        mappedSSHPort.Port(),
		RootPublicKey:  rootKeyPair.PublicKey,
		RootPrivateKey: rootKeyPair.PrivateKey,
	}

	if err = dc.ConfigureSSHD(ctx); err != nil {
		return nil, maybeTerminateContainerAfterError(ctx, container, err)
	}

	if err = dc.RestartSSHD(ctx); err != nil {
		return nil, maybeTerminateContainerAfterError(ctx, container, err)
	}
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
	exitCode, _, err := container.Exec(ctx, []string{"groupmod", "-g", "99", "systemd-timesync"})
	if exitCode != 0 {
		return fmt.Errorf("failed to change gid of containerized systemd-timesync group, got exit code %d\n", exitCode)
	} else if err != nil {
		return err
	}

	exitCode, _, err = container.Exec(ctx, []string{"groupmod", "-g", "101", "docker"})
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
