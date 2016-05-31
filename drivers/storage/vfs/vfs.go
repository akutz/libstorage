package vfs

import (
	"path"

	"github.com/akutz/gofig"
	"github.com/emccode/libstorage/api/types"
)

const (
	// Name is the name of the driver.
	Name = "vfs"
)

func init() {
	registerConfig()
}

func registerConfig() {
	defaultRootDir := types.Lib.Join("vfs")
	r := gofig.NewRegistration("VFS")
	r.Key(gofig.String, "", defaultRootDir, "", "vfs.root")
	gofig.Register(r)
}

// RootDir returns the path to the VFS root directory.
func RootDir(config gofig.Config) string {
	return config.GetString("vfs.root")
}

// DevicesDirPath returns the path to the VFS devices directory.
func DevicesDirPath(config gofig.Config) string {
	return path.Join(RootDir(config), "dev")
}

// DevicesFilePath returns the path to the VFS devices file.
func DevicesFilePath(config gofig.Config) string {
	return path.Join(RootDir(config), "dev.json")
}

// VolumesDirPath returns the path to the VFS volumes directory.
func VolumesDirPath(config gofig.Config) string {
	return path.Join(RootDir(config), "vol")
}

// SnapshotsDirPath returns the path to the VFS volumes directory.
func SnapshotsDirPath(config gofig.Config) string {
	return path.Join(RootDir(config), "snap")
}
