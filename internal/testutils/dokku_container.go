package testutils

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/ssh"
)

const (
	testKeyPath               = "/home/dokku/test_key.pub"
	testKeyFileMode           = 0666
	buildpackInstallScriptURL = "https://github.com/buildpacks/pack/releases/download/v0.27.0/pack-v0.27.0-linux.tgz"
)

type DokkuContainer struct {
	testcontainers.Container
	Host           string
	SSHPort        string
	logger         *logAccepter
	RootPrivateKey *rsa.PrivateKey
	RootPublicKey  []byte
}

type logAccepter struct {
	printFunc func(string, ...any)
}

func (la *logAccepter) Accept(l testcontainers.Log) {
	la.printFunc(string(l.Content))
}

func (dc *DokkuContainer) Cleanup(ctx context.Context) {
	dc.Terminate(ctx)
	if dc.logger != nil {
		dc.StopLogProducer()
	}
}

func (dc *DokkuContainer) InstallBuildPacksCLI(ctx context.Context) error {
	response, e := http.Get(buildpackInstallScriptURL)
	if e != nil {
		log.Fatal(e)
	}
	// todo: errors
	defer response.Body.Close()
	cliBytes, _ := ioutil.ReadAll(response.Body)
	dc.CopyToContainer(ctx, cliBytes, "/home/dokku/pack.tgz", 0666)
	installCmd := []string{"/usr/bin/tar",
		"-C", "/usr/local/bin/", "--no-same-owner", "-xzv", "pack", "-f", "/home/dokku/pack.tgz"}

	code, err := dc.Exec(ctx, installCmd)
	if err != nil {
		return fmt.Errorf("failed to install buildpacks: %w", err)
	} else if code != 0 {
		return fmt.Errorf("failed to install buildpacks: got exit code %d", code)
	}

	return nil
}

func (dc *DokkuContainer) AttachTestLogger(ctx context.Context, tb testing.TB) error {
	dc.logger = &logAccepter{printFunc: tb.Logf}

	if err := dc.StartLogProducer(ctx); err != nil {
		return err
	}

	dc.FollowOutput(dc.logger)

	return nil
}

func (dc *DokkuContainer) HostKeyFunc() func(string, net.Addr, ssh.PublicKey) error {
	return func(host string, remote net.Addr, key ssh.PublicKey) error {
		if net.JoinHostPort(dc.Host, dc.SSHPort) != host {
			return errors.New("invalid host supplied for handshake?")
		}
		return nil
	}
}

// func (dc *DokkuContainer) ConfigureSSHD(ctx context.Context) error {
// 	retCode, err := dc.Exec(ctx, []string{"sudo", "sed", "-i", "s/\\#PermitRootLogin yes/PermitRootLogin prohibit-password/", "/etc/ssh/sshd_config"})
// 	if err != nil {
// 		return fmt.Errorf("failed to configure sshd: %w", err)
// 	}
// 	if retCode != 0 {
// 		return fmt.Errorf("failed to configure sshd: got exit code %d", retCode)
// 	}
// 	return nil
// }

// func (dc *DokkuContainer) RestartSSHD(ctx context.Context) error {
// 	retCode, err := dc.Exec(ctx, []string{"service", "ssh", "restart"})
// 	if err != nil {
// 		return fmt.Errorf("failed to restart sshd: %w", err)
// 	}
// 	if retCode != 0 {
// 		return fmt.Errorf("failed to restart sshd: got exit code %d", retCode)
// 	}
// 	return nil
// }

func (dc *DokkuContainer) ExecAsRoot(ctx context.Context, cmd []string) (int, error) {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return 0, fmt.Errorf("failed to create docker client: %w", err)
	}
	execConfig := types.ExecConfig{User: "root", Cmd: cmd, AttachStdout: true, AttachStderr: true}
	execID, err := client.ContainerExecCreate(ctx, dc.GetContainerID(), execConfig)
	if err != nil {
		return 0, fmt.Errorf("failed to create exec: %w", err)
	}

	// resp, err := client.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to attach to exec: %w", err)
	// }
	// defer resp.Close()
	// go func() {
	// 	scanner := bufio.NewScanner(resp.Reader)
	// 	for scanner.Scan() {
	// 		dc.logger.Accept(testcontainers.Log{Content: scanner.Bytes()})
	// 	}
	// }()

	// <-ctx.Done()
	execInspect, err := client.ContainerExecInspect(ctx, execID.ID)
	for execInspect.Running {
		execInspect, err = client.ContainerExecInspect(ctx, execID.ID)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to inspect exec: %w", err)
	}
	return execInspect.ExitCode, nil
}

func (dc *DokkuContainer) RegisterRootAuthorizedKey(ctx context.Context) error {
	publicKey, err := ssh.ParsePublicKey(dc.RootPublicKey)
	if err != nil {
		return fmt.Errorf("failed to parse root public key: %w", err)
	}
	err = dc.CopyToContainer(ctx, ssh.MarshalAuthorizedKey(publicKey), "/root/.ssh/authorized_keys", 0666)
	if err != nil {
		return err
	}
	return nil
}

func (dc *DokkuContainer) RegisterPublicKey(ctx context.Context, key []byte, name string) error {
	err := dc.CopyToContainer(ctx, key, testKeyPath, testKeyFileMode)
	if err != nil {
		return err
	}

	chownCmd := []string{"/usr/bin/dokku", "ssh-keys:add", name, testKeyPath}
	retCode, err := dc.Exec(ctx, chownCmd)
	if err != nil {
		return fmt.Errorf("failed to add ssh key: %w", err)
	} else if retCode != 0 {
		return fmt.Errorf("failed to add ssh key: got exit code %d", retCode)
	}

	return nil
}
