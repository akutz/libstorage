package types

// define the integration config parent keys
const (
	_ ConfigKey = configSectionMaxParentsCommon - 1 + iota

	// ConfigIG is a config key.
	ConfigIG

	// ConfigIGVol is a config key.
	ConfigIGVol

	// ConfigIGVolOps is a config key.
	ConfigIGVolOps

	// ConfigIGVolOpsMount is a config key.
	ConfigIGVolOpsMount

	// ConfigIGVolOpsUnmount is a config key.
	ConfigIGVolOpsUnmount

	// ConfigIGVolOpsPath is a config key.
	ConfigIGVolOpsPath

	// ConfigIGVolOpsCreate is a config key.
	ConfigIGVolOpsCreate

	// ConfigIGVolOpsCreateDefault is a config key.
	ConfigIGVolOpsCreateDefault

	// ConfigIGVolOpsRemove is a config key.
	ConfigIGVolOpsRemove

	configSectionMaxParentsIG
)

// define the IG section keys
const (
	configSectionMinIG = configSectionMaxValsCommon - 1 + iota

	// ConfigIGDriver is a config key.
	ConfigIGDriver

	configSectionMaxIG
)

// define the IG Vol Ops Mount section keys
const (
	configSectionMinIGVolOpsMount = configSectionMaxIG - 1 + iota

	// ConfigIGVolOpsMountPreempt is a config key.
	ConfigIGVolOpsMountPreempt

	// ConfigIGVolOpsMountPath is a config key.
	ConfigIGVolOpsMountPath

	// ConfigIGVolOpsMountRootPath is a config key.
	ConfigIGVolOpsMountRootPath

	configSectionMaxIGVolOpsMount
)

// define the IG Vol Ops Unmount section keys
const (
	configSectionMinIGVolOpsUnmount = configSectionMaxIGVolOpsMount - 1 + iota

	// ConfigIGVolOpsUnmountIgnoreUsed is a config key.
	ConfigIGVolOpsUnmountIgnoreUsed

	configSectionMaxIGVolOpsUnmount
)

// define the IG Vol Ops Path section keys
const (
	configSectionMinIGVolOpsPath = configSectionMaxIGVolOpsUnmount - 1 + iota

	// ConfigIGVolOpsPathCache is a config key.
	ConfigIGVolOpsPathCache

	configSectionMaxIGVolOpsPath
)

// define the IG Vol Ops Create section keys
const (
	configSectionMinIGVolOpsCreate = configSectionMaxIGVolOpsPath - 1 + iota

	// ConfigIGVolOpsCreateDisable is a config key.
	ConfigIGVolOpsCreateDisable

	// ConfigIGVolOpsCreateImplicit is a config key.
	ConfigIGVolOpsCreateImplicit

	configSectionMaxIGVolOpsCreate
)

// define the IG Vol Ops Create Defaults section keys
const (
	configSectionMinIGVolOpsCreateDefault = configSectionMaxIGVolOpsCreate - 1 + iota

	// ConfigIGVolOpsCreateDefaultSize is a config key.
	ConfigIGVolOpsCreateDefaultSize

	// ConfigIGVolOpsCreateDefaultFSType is a config key.
	ConfigIGVolOpsCreateDefaultFSType

	// ConfigIGVolOpsCreateDefaultAZ is a config key.
	ConfigIGVolOpsCreateDefaultAZ

	// ConfigIGVolOpsCreateDefaultType is a config key.
	ConfigIGVolOpsCreateDefaultType

	// ConfigIGVolOpsCreateDefaultIOPS is a config key.
	ConfigIGVolOpsCreateDefaultIOPS

	configSectionMaxIGVolOpsCreateDefault
)

// define the IG Vol Ops Remove section keys
const (
	configSectionMinIGVolOpsRemove = configSectionMaxIGVolOpsCreateDefault - 1 + iota

	// ConfigIGVolOpsRemoveDisable is a config key.
	ConfigIGVolOpsRemoveDisable

	configSectionMaxIGVolOpsRemove
)

const configSectionMaxValsIG = configSectionMaxIGVolOpsRemove

const (
	configIG                          = configRoot + ".integration"
	configIGDriver                    = configIG + ".driver"
	configIGVol                       = configIG + ".volume"
	configIGVolOps                    = configIGVol + ".operations"
	configIGVolOpsMount               = configIGVolOps + ".mount"
	configIGVolOpsMountPreempt        = configIGVolOpsMount + ".preempt"
	configIGVolOpsMountPath           = configIGVolOpsMount + ".path"
	configIGVolOpsMountRootPath       = configIGVolOpsMount + ".rootPath"
	configIGVolOpsUnmount             = configIGVolOps + ".unmount"
	configIGVolOpsUnmountIgnoreUsed   = configIGVolOpsUnmount + ".ignoreusedcount"
	configIGVolOpsPath                = configIGVolOps + ".path"
	configIGVolOpsPathCache           = configIGVolOpsPath + ".cache"
	configIGVolOpsCreate              = configIGVolOps + ".create"
	configIGVolOpsCreateDisable       = configIGVolOpsCreate + ".disable"
	configIGVolOpsCreateImplicit      = configIGVolOpsCreate + ".implicit"
	configIGVolOpsCreateDefault       = configIGVolOpsCreate + ".default"
	configIGVolOpsCreateDefaultSize   = configIGVolOpsCreateDefault + ".size"
	configIGVolOpsCreateDefaultFSType = configIGVolOpsCreateDefault + ".fsType"
	configIGVolOpsCreateDefaultAZ     = configIGVolOpsCreateDefault + ".availabilityZone"
	configIGVolOpsCreateDefaultType   = configIGVolOpsCreateDefault + ".type"
	configIGVolOpsCreateDefaultIOPS   = configIGVolOpsCreateDefault + ".IOPS"
	configIGVolOpsRemove              = configIGVolOps + ".remove"
	configIGVolOpsRemoveDisable       = configIGVolOpsRemove + ".disable"
)

var configIGKeyPaths = map[ConfigKey]string{
	ConfigIG:                          configIG,
	ConfigIGDriver:                    configIGDriver,
	ConfigIGVol:                       configIGVol,
	ConfigIGVolOps:                    configIGVolOps,
	ConfigIGVolOpsMount:               configIGVolOpsMount,
	ConfigIGVolOpsMountPreempt:        configIGVolOpsMountPreempt,
	ConfigIGVolOpsMountPath:           configIGVolOpsMountPath,
	ConfigIGVolOpsMountRootPath:       configIGVolOpsMountRootPath,
	ConfigIGVolOpsUnmount:             configIGVolOpsUnmount,
	ConfigIGVolOpsUnmountIgnoreUsed:   configIGVolOpsUnmountIgnoreUsed,
	ConfigIGVolOpsPath:                configIGVolOpsPath,
	ConfigIGVolOpsPathCache:           configIGVolOpsPathCache,
	ConfigIGVolOpsCreate:              configIGVolOpsCreate,
	ConfigIGVolOpsCreateDisable:       configIGVolOpsCreateDisable,
	ConfigIGVolOpsCreateImplicit:      configIGVolOpsCreateImplicit,
	ConfigIGVolOpsCreateDefault:       configIGVolOpsCreateDefault,
	ConfigIGVolOpsCreateDefaultSize:   configIGVolOpsCreateDefaultSize,
	ConfigIGVolOpsCreateDefaultFSType: configIGVolOpsCreateDefaultFSType,
	ConfigIGVolOpsCreateDefaultAZ:     configIGVolOpsCreateDefaultAZ,
	ConfigIGVolOpsCreateDefaultType:   configIGVolOpsCreateDefaultType,
	ConfigIGVolOpsCreateDefaultIOPS:   configIGVolOpsCreateDefaultIOPS,
	ConfigIGVolOpsRemove:              configIGVolOpsRemove,
	ConfigIGVolOpsRemoveDisable:       configIGVolOpsRemoveDisable,
}

var configIGKeyDescs = map[ConfigKey]string{}
