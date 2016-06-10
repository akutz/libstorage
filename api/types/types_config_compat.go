package types

import (
	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
)

const (
	coldIGRoot                 = "volume"
	coldIGVolMountPreempt      = coldIGRoot + ".mount.preempt"
	coldIGVolCreateDisable     = coldIGRoot + ".create.disable"
	coldIGVolRemoveDisable     = coldIGRoot + ".remove.disable"
	coldIGVolUnmountIgnoreUsed = coldIGRoot + ".unmount.ignoreusedcount"
	coldIGVolPathCache         = coldIGRoot + ".path.cache"
	coldDockerRoot             = "docker"
	coldDockerFsType           = coldDockerRoot + ".fsType"
	coldDockerVolumeType       = coldDockerRoot + ".volumeType"
	coldDockerIOPS             = coldDockerRoot + ".iops"
	coldDockerSize             = coldDockerRoot + ".size"
	coldDockerAvailabilityZone = coldDockerRoot + ".availabilityZone"
	coldDockerMountDirPath     = coldDockerRoot + ".mountDirPath"
	coldLinuxVolumeRootPath    = "linux.volume.rootpath"
)

// BackCompat ensures keys can be used from old configurations.
func BackCompat(ctx Context, config gofig.Config) {
	keyMap := map[ConfigKey]string{
		ConfigIGVolOpsMountPreempt:        coldIGVolMountPreempt,
		ConfigIGVolOpsCreateDisable:       coldIGVolCreateDisable,
		ConfigIGVolOpsRemoveDisable:       coldIGVolRemoveDisable,
		ConfigIGVolOpsUnmountIgnoreUsed:   coldIGVolUnmountIgnoreUsed,
		ConfigIGVolOpsPathCache:           coldIGVolPathCache,
		ConfigIGVolOpsCreateDefaultFSType: coldDockerFsType,
		ConfigIGVolOpsCreateDefaultType:   coldDockerVolumeType,
		ConfigIGVolOpsCreateDefaultIOPS:   coldDockerIOPS,
		ConfigIGVolOpsCreateDefaultSize:   coldDockerSize,
		ConfigIGVolOpsCreateDefaultAZ:     coldDockerAvailabilityZone,
		ConfigIGVolOpsMountPath:           coldDockerMountDirPath,
		ConfigIGVolOpsMountRootPath:       coldLinuxVolumeRootPath,
	}
	for newKey, oldKey := range keyMap {
		if !config.IsSet(newKey) && config.IsSet(oldKey) {
			val := config.Get(oldKey)
			ctx.WithFields(log.Fields{
				"oldKey": oldKey,
				"newKey": newKey,
				"val":    val,
			}).Debug("setting val from old key on new key")
			config.Set(newKey, val)
		}
	}
}
