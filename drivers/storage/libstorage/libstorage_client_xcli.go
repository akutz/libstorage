package libstorage

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func (c *client) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	if c.isController() {
		return nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "InstanceID")
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return nil, goof.New("missing service name")
	}

	if iid := c.instanceIDCache.GetInstanceID(serviceName); iid != nil {
		return iid, nil
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	driverName := strings.ToLower(si.Driver.Name)

	out, err := c.runExecutor(ctx, driverName, types.LSXCmdInstanceID)
	if err != nil {
		return nil, err
	}

	iid := &types.InstanceID{}
	if err := iid.UnmarshalText(out); err != nil {
		return nil, err
	}

	ctx = ctx.WithValue(context.InstanceIDKey, iid)

	ctx.Debug("sending instanceID in API.InstanceInspect call")
	instance, err := c.InstanceInspect(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	iid.ID = instance.InstanceID.ID
	iid.DeleteMetadata()
	c.instanceIDCache.Set(serviceName, iid)
	ctx.Debug("received instanceID from API.InstanceInspect call")

	ctx.Debug("xli instanceID success")
	return iid, nil
}

func (c *client) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	if c.isController() {
		return "", utils.NewUnsupportedForClientTypeError(
			c.clientType, "NextDevice")
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return "", goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return "", err
	}
	driverName := si.Driver.Name

	out, err := c.runExecutor(ctx, driverName, types.LSXCmdNextDevice)
	if err != nil {
		return "", err
	}

	ctx.Debug("xli nextdevice success")
	return gotil.Trim(string(out)), nil
}

func (c *client) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	if c.isController() {
		return nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "LocalDevices")
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return nil, goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	driverName := si.Driver.Name

	out, err := c.runExecutor(
		ctx, driverName, types.LSXCmdLocalDevices, opts.ScanType.String())
	if err != nil {
		return nil, err
	}

	ld := &types.LocalDevices{}
	if err := ld.UnmarshalText(out); err != nil {
		return nil, err
	}

	ctx.Debug("xli localdevices success")
	return ld, nil
}

func (c *client) WaitForDevice(
	ctx types.Context,
	opts *types.WaitForDeviceOpts) (bool, *types.LocalDevices, error) {

	if c.isController() {
		return false, nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "WaitForDevice")
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return false, nil, goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return false, nil, err
	}
	driverName := si.Driver.Name

	exitCode := 0
	out, err := c.runExecutor(
		ctx, driverName, types.LSXCmdWaitForDevice,
		opts.ScanType.String(), opts.Token, opts.Timeout.String())
	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode = exitError.Sys().(syscall.WaitStatus).ExitStatus()
	}

	if err != nil && exitCode > 0 {
		return false, nil, err
	}

	matched := exitCode == 0

	ld := &types.LocalDevices{}
	if err := ld.UnmarshalText(out); err != nil {
		return false, nil, err
	}

	ctx.Debug("xli waitfordevice success")
	return matched, ld, nil
}

// MapDevice creates a mapping between a volume ID and a device path.
func (c *client) MapDevice(
	ctx types.Context,
	volumeID, devicePath string,
	opts types.Store) error {

	if c.isController() {
		return utils.NewUnsupportedForClientTypeError(
			c.clientType, "MapDevice")
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return err
	}
	driverName := si.Driver.Name

	exitCode := 0
	_, err = c.runExecutor(
		ctx, driverName, types.LSXCmdMapDevice,
		volumeID, devicePath)
	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode = exitError.Sys().(syscall.WaitStatus).ExitStatus()
	}

	if err != nil && exitCode > 0 {
		return err
	}

	ctx.Debug("xli mapdevice success")
	return nil
}

func (c *client) runExecutor(
	ctx types.Context, args ...string) ([]byte, error) {

	if c.isController() {
		return nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "runExecutor")
	}

	ctx.Debug("waiting on executor lock")
	if err := c.lsxMutexWait(ctx); err != nil {
		return nil, err
	}

	defer func() {
		ctx.Debug("signalling executor lock")
		if err := c.lsxMutexSignal(ctx); err != nil {
			panic(err)
		}
	}()

	cmd := exec.Command(types.LSX.String(), args...)
	cmd.Env = os.Environ()

	configEnvVars := c.config.EnvVars()
	for _, cev := range configEnvVars {
		// ctx.WithField("value", cev).Debug("set executor env var")
		cmd.Env = append(cmd.Env, cev)
	}

	return cmd.Output()
}

func (c *client) lsxMutexWait(ctx types.Context) error {

	if c.isController() {
		return utils.NewUnsupportedForClientTypeError(
			c.clientType, "lsxMutexWait")
	}

	lsxMutex.Lock()
	return nil
}

func (c *client) lsxMutexSignal(ctx types.Context) error {
	if c.isController() {
		return utils.NewUnsupportedForClientTypeError(
			c.clientType, "lsxMutexSignal")
	}

	lsxMutex.Unlock()
	return nil
}
