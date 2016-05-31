package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/drivers/storage/vfs"
)

type driver struct {
	config      gofig.Config
	rootDir     string
	devDirPath  string
	devFilePath string
	osDriver    types.OSDriver
}

func init() {
	registry.RegisterStorageExecutor(vfs.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Name() string {
	return vfs.Name
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {

	d.config = config
	d.rootDir = vfs.RootDir(config)
	d.devDirPath = vfs.DevicesDirPath(config)

	if err := os.MkdirAll(d.rootDir, 0755); err != nil {
		return err
	}

	d.devFilePath = vfs.DevicesFilePath(config)
	d.initLocalDevices()

	osDriver, err := registry.NewOSDriver(runtime.GOOS)
	if err != nil {
		return err
	}
	if err := osDriver.Init(ctx, config); err != nil {
		return err
	}
	d.osDriver = osDriver

	return nil
}

// InstanceID returns the local system's InstanceID.
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	hostName, err := utils.HostName()
	if err != nil {
		return nil, err
	}

	iid := &types.InstanceID{Driver: vfs.Name}

	if err := iid.MarshalMetadata(hostName); err != nil {
		return nil, err
	}

	return iid, nil
}

var intA = int('a')

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	devMap, devs, err := d.readLocalDevices()
	if err != nil {
		return "", err
	}

	if len(devMap) == 0 {
		devPath := path.Join(d.devDirPath, "xvda")
		ctx.WithField("path", devPath).Debug("initial device")
		return devPath, nil
	}

	for _, devicePath := range devs {

		devName := fmt.Sprintf("bind@/dev/%s", path.Base(devicePath))

		mountInfos, err := d.osDriver.Mounts(ctx, devName, "", opts)
		if err != nil {
			return "", nil
		}
		// if the device is not presently mounted then return it
		if len(mountInfos) == 0 {
			ctx.WithField("path", devicePath).Debug("available device")
			return devicePath, nil
		}
	}

	var (
		lastDevPath  = devs[len(devs)-1]
		lastDevPathC = lastDevPath[len(lastDevPath)-1]
		nextDevPathC = lastDevPathC + 1
		nextDevName  = fmt.Sprintf("xvd%c", nextDevPathC)
		nextDevPath  = path.Join(d.devDirPath, nextDevName)
	)

	ctx.WithField("path", nextDevPath).Debug("next device")
	return nextDevPath, nil
}

var (
	devRX = regexp.MustCompile(`^(/dev/xvd[a-z])(?:=(.+))?$`)
)

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	ctx.WithFields(log.Fields{
		"vfs.root": d.rootDir,
		"dev.path": d.devDirPath,
	}).Debug("config info")

	devMap, _, err := d.readLocalDevices()
	if err != nil {
		return nil, err
	}

	return &types.LocalDevices{Driver: vfs.Name, DeviceMap: devMap}, nil
}

// MapDevice creates a mapping between a volume ID and a device path.
func (d *driver) MapDevice(
	ctx types.Context,
	volumeID, devicePath string,
	opts types.Store) error {

	ctx.WithFields(log.Fields{
		"vfs.root":   d.rootDir,
		"volumeID":   volumeID,
		"devicePath": devicePath,
	}).Debug("map device")

	devMap, _, err := d.readLocalDevices()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(devicePath, 0755); err != nil {
		return err
	}

	devMap[volumeID] = devicePath
	if err := d.writeLocalDevices(devMap); err != nil {
		return err
	}

	return nil
}

func (d *driver) getDevDirs() ([]string, error) {
	if !gotil.FileExists(d.devDirPath) {
		return nil, goof.WithField(
			"path", d.devDirPath, "invalid device dir path")
	}

	matches, err := filepath.Glob(path.Join(d.devDirPath, "xvd*"))
	if err != nil {
		return nil, err
	}

	return matches, nil
}

var (
	devFileLock = types.NewLockFile(types.Run.Join("lsx.vfs.lock"))
)

func (d *driver) readLocalDevices() (map[string]string, []string, error) {

	devMap := map[string]string{}
	devices := []string{}

	devFileLock.Lock()
	defer devFileLock.Unlock()

	f, err := os.Open(d.devFilePath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&devMap); err != nil {
		return nil, nil, err
	}

	for _, d := range devMap {
		devices = append(devices, d)
	}

	// sort the devices lexographically
	devices = utils.SortByString(devices)

	return devMap, devices, nil
}

func (d *driver) writeLocalDevices(devMap map[string]string) error {

	devFileLock.Lock()
	defer devFileLock.Unlock()

	f, err := os.Create(d.devFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(devMap); err != nil {
		return err
	}

	return nil
}

func (d *driver) initLocalDevices() error {

	devFileLock.Lock()
	defer devFileLock.Unlock()

	if gotil.FileExists(d.devFilePath) {
		return nil
	}

	f, err := os.Create(d.devFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]string{}); err != nil {
		return err
	}

	return nil
}
