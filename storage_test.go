package dokku

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type storageManagerTestSuite struct {
	dokkuTestSuite
}

func TestRunStorageManagerTestSuite(t *testing.T) {
	suite.Run(t, new(storageManagerTestSuite))
}

func (s *storageManagerTestSuite) TestManageStorage() {
	r := s.Suite.Require()

	appName := "test-storage-app"

	r.NoError(s.Client.CreateApp(appName))

	storageReport, err := s.Client.GetAppStorageReport(appName)
	r.NoError(err)
	r.Len(storageReport.RunMounts, 0)

	storage := StorageBindMount{
		HostDir:      "testAppStorage",
		ContainerDir: "/data",
	}
	r.NoError(s.Client.EnsureStorageDirectory(storage.HostDir, StorageChownOptionHerokuish))
	r.NoError(s.Client.MountAppStorage(appName, storage))

	storageList, err := s.Client.ListAppStorage(appName)
	r.NoError(err)
	r.Len(storageList, 1)
	r.Equal(storage, storageList[0])

	storage2 := StorageBindMount{
		HostDir:      "testAppStorage2",
		ContainerDir: "/data2",
	}
	r.NoError(s.Client.EnsureStorageDirectory(storage2.HostDir, StorageChownOptionHerokuish))
	r.NoError(s.Client.MountAppStorage(appName, storage2))

	storageReport, err = s.Client.GetAppStorageReport(appName)
	r.NoError(err)
	r.Contains(storageReport.RunMounts, storage2)
}
