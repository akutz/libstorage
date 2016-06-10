package types

// ConfigKey is a configuration key.
type ConfigKey uint32

const (
	// ConfigKeyUnknown is an unknown configuration key.
	ConfigKeyUnknown ConfigKey = 0
)

// define the parent keys
const (
	_ ConfigKey = iota

	// ConfigRoot is a config key.
	ConfigRoot

	// ConfigClient is a config key.
	ConfigClient

	// ConfigClientCache is a config key.
	ConfigClientCache

	// ConfigServer is a config key.
	ConfigServer

	// ConfigLogging is a config key.
	ConfigLogging

	// ConfigServices is a config key.
	ConfigServices

	// ConfigEndpoints is a config key.
	ConfigEndpoints

	// ConfigTLS is a config key.
	ConfigTLS

	// ConfigHTTP is a config key.
	ConfigHTTP

	// ConfigExecutor is a config key.
	ConfigExecutor

	// ConfigDevice is a config key.
	ConfigDevice

	// ConfigOS is a config key.
	ConfigOS

	// ConfigStorage is a config key.
	ConfigStorage

	configSectionMaxParentsCommon
)

const configKeyValOffset = 1024

// define the root section keys
const (
	configSectionMinRoot ConfigKey = configKeyValOffset + iota

	// ConfigHost is a config key.
	ConfigHost

	// ConfigEmbedded is a config key.
	ConfigEmbedded

	// ConfigService is a config key.
	ConfigService

	configSectionMaxRoot
)

// define the server section keys
const (
	configSectionMinServer = configSectionMaxRoot - 1 + iota

	// ConfigServerAutoEndpointMode is a config key.
	ConfigServerAutoEndpointMode

	configSectionMaxServer
)

// define the client section keys
const (
	configSectionMinClient = configSectionMaxServer - 1 + iota

	// ConfigClientType is a config key.
	ConfigClientType

	configSectionMaxClient
)

// define the client cache section keys
const (
	configSectionMinClientCache = configSectionMaxClient - 1 + iota

	// ConfigClientCacheInstanceID is a config key.
	ConfigClientCacheInstanceID

	configSectionMaxClientCache
)

// define the OS section keys
const (
	configSectionMinOS = configSectionMaxClientCache - 1 + iota

	// ConfigOSDriver is a config key.
	ConfigOSDriver

	configSectionMaxOS
)

// define the storage section keys
const (
	configSectionMinStorage = configSectionMaxOS - 1 + iota

	// ConfigStorageDriver is a config key.
	ConfigStorageDriver

	configSectionMaxStorage
)

// define the logging section keys
const (
	configSectionMinLogging = configSectionMaxStorage - 1 + iota

	// ConfigLogLevel is a config key.
	ConfigLogLevel

	// ConfigLogStdout is a config key.
	ConfigLogStdout

	// ConfigLogStderr is a config key.
	ConfigLogStderr

	// ConfigLogHTTPRequests is a config key.
	ConfigLogHTTPRequests

	// ConfigLogHTTPResponses is a config key.
	ConfigLogHTTPResponses

	configSectionMaxLogging
)

// define the http section keys
const (
	configSectionMinHTTP = configSectionMaxLogging - 1 + iota

	// ConfigHTTPDisableKeepAlive is a config key.
	ConfigHTTPDisableKeepAlive

	// ConfigHTTPWriteTimeout is a config key.
	ConfigHTTPWriteTimeout

	// ConfigHTTPReadTimeout is a config key.
	ConfigHTTPReadTimeout

	configSectionMaxHTTP
)

// define the executor section keys
const (
	configSectionMinExecutor = configSectionMaxHTTP - 1 + iota

	// ConfigExecutorPath is a config key.
	ConfigExecutorPath

	// ConfigExecutorNoDownload is a config key.
	ConfigExecutorNoDownload

	configSectionMaxExecutor
)

// define the TLS section keys
const (
	configSectionMinTLS = configSectionMaxExecutor - 1 + iota

	// ConfigTLSDisabled is a config key.
	ConfigTLSDisabled

	// ConfigTLSServerName is a config key.
	ConfigTLSServerName

	// ConfigTLSClientCertRequired is a config key.
	ConfigTLSClientCertRequired

	// ConfigTLSTrustedCertsFile is a config key.
	ConfigTLSTrustedCertsFile

	// ConfigTLSCertFile is a config key.
	ConfigTLSCertFile

	// ConfigTLSKeyFile is a config key.
	ConfigTLSKeyFile

	configSectionMaxTLS
)

// define the device section keys
const (
	configSectionMinDevice = configSectionMaxTLS - 1 + iota

	// ConfigDeviceAttachTimeout is a config key.
	ConfigDeviceAttachTimeout

	// ConfigDeviceScanType is a config key.
	ConfigDeviceScanType

	configSectionMaxDevice
)

const (
	configSectionMaxValsCommon = configSectionMaxDevice

	// configSectionMaxParents is the max key from all parent config keys
	configSectionMaxParents = configSectionMaxParentsIG

	// configSectionMaxVals is the max key from all value config keys
	configSectionMaxVals = configSectionMaxValsIG
)

var (
	parentConfigKeys [configSectionMaxParents - 1]ConfigKey
	valueConfigKeys  [configSectionMaxVals - configKeyValOffset - 1]ConfigKey
)

func initConfigKeys() {

	for x := 0; x < len(parentConfigKeys); x++ {
		parentConfigKeys[x] = ConfigKey(x + 1)
	}

	for x := 0; x < len(valueConfigKeys); x++ {
		valueConfigKeys[x] = ConfigKey(x + 1 + configKeyValOffset)
	}
}

// ParentConfigKeys returns a channel on which can be received all the parent
// configuration keys.
func ParentConfigKeys() <-chan ConfigKey {
	return newConfigKeyChan(parentConfigKeys[:])
}

// ValueConfigKeys returns a channel on which can be received all the value
// configuration keys.
func ValueConfigKeys() <-chan ConfigKey {
	return newConfigKeyChan(valueConfigKeys[:])
}

func newConfigKeyChan(keys []ConfigKey) <-chan ConfigKey {
	c := make(chan ConfigKey)
	go func() {
		for _, k := range keys {
			c <- k
		}
		close(c)
	}()
	return c
}

// String returns the ConfigKey's path.
func (k ConfigKey) String() string {
	if v, ok := configCommonKeyPaths[k]; ok {
		return v
	}
	if v, ok := configIGKeyPaths[k]; ok {
		return v
	}
	return ""
}

const (
	configRoot                   = "libstorage"
	configServer                 = configRoot + ".server"
	configClient                 = configRoot + ".client"
	configClientType             = configRoot + ".type"
	configHost                   = configRoot + ".host"
	configEmbedded               = configRoot + ".embedded"
	configService                = configRoot + ".service"
	configOS                     = configRoot + ".os"
	configStorage                = configRoot + ".storage"
	configOSDriver               = configOS + ".driver"
	configStorageDriver          = configStorage + ".driver"
	configLogging                = configRoot + ".logging"
	configLogLevel               = configLogging + ".level"
	configLogStdout              = configLogging + ".stdout"
	configLogStderr              = configLogging + ".stderr"
	configLogHTTPRequests        = configLogging + ".httpRequests"
	configLogHTTPResponses       = configLogging + ".httpResponses"
	configHTTP                   = configRoot + ".http"
	configHTTPDisableKeepAlive   = configHTTP + ".disableKeepAlive"
	configHTTPWriteTimeout       = configHTTP + ".writeTimeout"
	configHTTPReadTimeout        = configHTTP + ".readTimeout"
	configServices               = configServer + ".services"
	configServerAutoEndpointMode = configServer + ".autoEndpointMode"
	configEndpoints              = configServer + ".endpoints"
	configExecutor               = configRoot + ".executor"
	configExecutorPath           = configExecutor + ".path"
	configExecutorNoDownload     = configExecutor + ".disableDownload"
	configClientCache            = configClient + ".cache"
	configClientCacheInstanceID  = configClientCache + ".instanceID"
	configTLS                    = configRoot + ".tls"
	configTLSDisabled            = configTLS + ".disabled"
	configTLSServerName          = configTLS + ".serverName"
	configTLSClientCertRequired  = configTLS + ".clientCertRequired"
	configTLSTrustedCertsFile    = configTLS + ".trustedCertsFile"
	configTLSCertFile            = configTLS + ".certFile"
	configTLSKeyFile             = configTLS + ".keyFile"
	configDevice                 = configRoot + ".device"
	configDeviceAttachTimeout    = configDevice + ".attachTimeout"
	configDeviceScanType         = configDevice + ".scanType"
)

var configCommonKeyPaths = map[ConfigKey]string{
	ConfigRoot:                   configRoot,
	ConfigServer:                 configServer,
	ConfigClient:                 configClient,
	ConfigClientType:             configClientType,
	ConfigHost:                   configHost,
	ConfigEmbedded:               configEmbedded,
	ConfigService:                configService,
	ConfigOS:                     configOS,
	ConfigStorage:                configStorage,
	ConfigOSDriver:               configOSDriver,
	ConfigStorageDriver:          configStorageDriver,
	ConfigLogging:                configLogging,
	ConfigLogLevel:               configLogLevel,
	ConfigLogStdout:              configLogStdout,
	ConfigLogStderr:              configLogStderr,
	ConfigLogHTTPRequests:        configLogHTTPRequests,
	ConfigLogHTTPResponses:       configLogHTTPResponses,
	ConfigHTTP:                   configHTTP,
	ConfigHTTPDisableKeepAlive:   configHTTPDisableKeepAlive,
	ConfigHTTPWriteTimeout:       configHTTPWriteTimeout,
	ConfigHTTPReadTimeout:        configHTTPReadTimeout,
	ConfigServices:               configServices,
	ConfigServerAutoEndpointMode: configServerAutoEndpointMode,
	ConfigEndpoints:              configEndpoints,
	ConfigExecutor:               configExecutor,
	ConfigExecutorPath:           configExecutorPath,
	ConfigExecutorNoDownload:     configExecutorNoDownload,
	ConfigClientCache:            configClientCache,
	ConfigClientCacheInstanceID:  configClientCacheInstanceID,
	ConfigTLS:                    configTLS,
	ConfigTLSDisabled:            configTLSDisabled,
	ConfigTLSServerName:          configTLSServerName,
	ConfigTLSClientCertRequired:  configTLSClientCertRequired,
	ConfigTLSTrustedCertsFile:    configTLSTrustedCertsFile,
	ConfigTLSCertFile:            configTLSCertFile,
	ConfigTLSKeyFile:             configTLSKeyFile,
	ConfigDevice:                 configDevice,
	ConfigDeviceAttachTimeout:    configDeviceAttachTimeout,
	ConfigDeviceScanType:         configDeviceScanType,
}

var configCommonKeyDescs = map[ConfigKey]string{}

// configKeyHierarchy is a map where the keys are config key section names that
// point to slices where the first element and second elements are the min and
// max value keys for that section, or 0 and 0 if that section has no vlaues.
// The remaining elements are the keys that are children of that section
var configKeyHierarchy = map[ConfigKey][]ConfigKey{
	ConfigRoot: []ConfigKey{
		configSectionMinRoot,
		configSectionMaxRoot,
		ConfigClient, ConfigServer, ConfigLogging, ConfigTLS, ConfigHTTP,
		ConfigExecutor, ConfigDevice, ConfigOS, ConfigStorage, ConfigIG,
	},
	ConfigClient: []ConfigKey{
		configSectionMinClient,
		configSectionMaxClient,
		ConfigClientCache,
	},
	ConfigClientCache: []ConfigKey{
		configSectionMinClientCache,
		configSectionMaxClientCache,
	},
	ConfigServer: []ConfigKey{
		configSectionMinServer,
		configSectionMaxServer,
	},
	ConfigLogging: []ConfigKey{
		configSectionMinLogging,
		configSectionMaxLogging,
	},
	ConfigTLS: []ConfigKey{
		configSectionMinTLS,
		configSectionMaxTLS,
	},
	ConfigHTTP: []ConfigKey{
		configSectionMinHTTP,
		configSectionMaxHTTP,
	},
	ConfigExecutor: []ConfigKey{
		configSectionMinExecutor,
		configSectionMaxExecutor,
	},
	ConfigDevice: []ConfigKey{
		configSectionMinDevice,
		configSectionMaxDevice,
	},
	ConfigOS: []ConfigKey{
		configSectionMinOS,
		configSectionMaxOS,
	},
	ConfigStorage: []ConfigKey{
		configSectionMinStorage,
		configSectionMaxStorage,
	},
	ConfigIG: []ConfigKey{
		configSectionMinIG,
		configSectionMaxIG,
		ConfigIGVolOpsCreate, ConfigIGVolOpsRemove, ConfigIGVolOpsMount,
		ConfigIGVolOpsUnmount, ConfigIGVolOpsPath,
	},
	ConfigIGVolOpsCreate: []ConfigKey{
		configSectionMinIGVolOpsCreate,
		configSectionMaxIGVolOpsCreate,
		ConfigIGVolOpsCreateDefault,
	},
	ConfigIGVolOpsCreateDefault: []ConfigKey{
		configSectionMinIGVolOpsCreateDefault,
		configSectionMaxIGVolOpsCreateDefault,
	},
	ConfigIGVolOpsRemove: []ConfigKey{
		configSectionMinIGVolOpsRemove,
		configSectionMaxIGVolOpsRemove,
	},
	ConfigIGVolOpsMount: []ConfigKey{
		configSectionMinIGVolOpsMount,
		configSectionMaxIGVolOpsMount,
	},
	ConfigIGVolOpsUnmount: []ConfigKey{
		configSectionMinIGVolOpsUnmount,
		configSectionMaxIGVolOpsUnmount,
	},
	ConfigIGVolOpsPath: []ConfigKey{
		configSectionMinIGVolOpsPath,
		configSectionMaxIGVolOpsPath,
	},
}

// GetConfigSectionInfo gets a ConfigKey's value keys and keys that point to
// any child sections.
func GetConfigSectionInfo(key ConfigKey) ([]ConfigKey, []ConfigKey, bool) {

	info, ok := configKeyHierarchy[key]
	if !ok {
		return nil, nil, false
	}

	var (
		minKey     = info[0]
		maxKey     = info[1]
		valKeysLen = maxKey - minKey - 1
		valKeys    = make([]ConfigKey, valKeysLen)
	)

	x := 0
	for k := minKey + 1; k < maxKey; k++ {
		valKeys[x] = k
		x++
	}

	if len(info) < 3 {
		return valKeys, nil, true
	}

	var (
		subKeysLen = len(info) - 2
		subKeys    = make([]ConfigKey, subKeysLen)
	)

	for x := 2; x < len(info); x++ {
		subKeys[x-2] = info[x]
	}

	return valKeys, subKeys, true
}
