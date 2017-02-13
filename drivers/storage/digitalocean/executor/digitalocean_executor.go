// +build !libstorage_storage_executor libstorage_storage_executor_digitalocean

package executor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	gofig "github.com/akutz/gofig/types"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
	do "github.com/codedellemc/libstorage/drivers/storage/digitalocean"
	doUtils "github.com/codedellemc/libstorage/drivers/storage/digitalocean/utils"
)

var (
	diskPrefix = regexp.MustCompile(
		fmt.Sprintf("^%s(.*)", do.VolumePrefix))
	diskSuffix = regexp.MustCompile("part-.*$")
)

type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(do.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Name() string {
	return do.Name
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) InstanceID(
	ctx types.Context, opts types.Store) (*types.InstanceID, error) {
	return doUtils.InstanceID(ctx)
}

func (d *driver) NextDevice(
	ctx types.Context, opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {
	deviceMap := map[string]string{}
	diskIDPath := "/dev/disk/by-id"

	dir, _ := ioutil.ReadDir(diskIDPath)
	for _, device := range dir {
		switch {
		case !diskPrefix.MatchString(device.Name()):
			continue
		case diskSuffix.MatchString(device.Name()):
			continue
		case diskPrefix.MatchString(device.Name()):
			volumeName := diskPrefix.FindStringSubmatch(device.Name())[1]
			devPath, err := filepath.EvalSymlinks(
				fmt.Sprintf("%s/%s", diskIDPath, device.Name()))
			if err != nil {
				return nil, err
			}
			deviceMap[volumeName] = devPath
		}
	}

	ld := &types.LocalDevices{Driver: d.Name()}
	if len(deviceMap) > 0 {
		ld.DeviceMap = deviceMap
	}

	return ld, nil
}

func (d *driver) Supported(ctx types.Context, opts types.Store) (bool, error) {
	iid, err := d.InstanceID(ctx, opts)
	if err != nil {
		return false, err
	}

	token := d.token()
	region := iid.Fields[do.InstanceIDFieldRegion]
	client, err := doUtils.Client(token)
	if err != nil {
		return false, err
	}

	regions, _, err := client.Regions.List(nil)
	if err != nil {
		return false, err
	}

	for _, reg := range regions {
		if reg.Slug == region {
			for _, feature := range reg.Features {
				if feature == "storage" {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (d *driver) token() string {
	return d.config.GetString(do.ConfigDOToken)
}
