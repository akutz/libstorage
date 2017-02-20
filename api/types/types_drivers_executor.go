package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DeviceScanType is a type of device scan algorithm.
type DeviceScanType int

const (
	// LSXExitCodeNotImplemented is the exit code the executor binary uses to
	// indicate a function is not implemented for a given storage driver on the
	// current system.
	LSXExitCodeNotImplemented = 2

	// LSXExitCodeTimedOut is the exit code the executor binary uses to indicate
	// a function timed out.
	LSXExitCodeTimedOut = 255

	// LSXCmdInstanceID is the command to execute to get the instance ID.
	LSXCmdInstanceID = "instanceID"

	// LSXCmdLocalDevices is the command to execute to get the local devices
	// map.
	LSXCmdLocalDevices = "localDevices"

	// LSXCmdNextDevice is the command to execute to get the next device.
	LSXCmdNextDevice = "nextDevice"

	// LSXCmdWaitForDevice is the command to execute to wait until a device,
	// identified by volume ID, is presented to the system.
	LSXCmdWaitForDevice = "wait"

	// LSXCmdSupported is the command to execute to find out if an executor
	// is valid for a given platform on the current host.
	LSXCmdSupported = "supported"

	// LSXCmdMount is the command for mounting a device to a file system path.
	LSXCmdMount = "mount"

	// LSXCmdUmount is the command for unmounting mounted file systems.
	LSXCmdUmount = "umount"

	// LSXCmdMounts is the command for getting a list of mount info objects.
	LSXCmdMounts = "mounts"

	// LSXCmdVolumeCreate is the command for creating a volume.
	LSXCmdVolumeCreate = "volumeCreate"

	// LSXCmdVolumeRemove is the command for removing a volume
	LSXCmdVolumeRemove = "volumeRemove"

	// LSXCmdVolumeAttach is the command for attaching a volume.
	LSXCmdVolumeAttach = "volumeAttach"

	// LSXCmdVolumeDetach is the command for detaching a volume.
	LSXCmdVolumeDetach = "volumeDetach"
)

const (

	// DeviceScanQuick performs a shallow, quick scan.
	DeviceScanQuick DeviceScanType = iota

	// DeviceScanDeep performs a deep, longer scan.
	DeviceScanDeep
)

// String returns the string representation of a DeviceScanType.
func (st DeviceScanType) String() string {
	switch st {
	case DeviceScanQuick:
		return "quick"
	case DeviceScanDeep:
		return "deep"
	}
	return ""
}

// ParseDeviceScanType parses a device scan type.
func ParseDeviceScanType(i interface{}) DeviceScanType {
	switch ti := i.(type) {
	case string:
		lti := strings.ToLower(ti)
		if lti == DeviceScanQuick.String() {
			return DeviceScanQuick
		} else if lti == DeviceScanDeep.String() {
			return DeviceScanDeep
		}
		i, err := strconv.Atoi(ti)
		if err != nil {
			return DeviceScanQuick
		}
		return ParseDeviceScanType(i)
	case int:
		st := DeviceScanType(ti)
		if st == DeviceScanQuick || st == DeviceScanDeep {
			return st
		}
		return DeviceScanQuick
	default:
		return ParseDeviceScanType(fmt.Sprintf("%v", ti))
	}
}

// LocalDevicesOpts are options when getting a list of local devices.
type LocalDevicesOpts struct {
	ScanType DeviceScanType
	Opts     Store
}

// WaitForDeviceOpts are options when waiting on specific local device to
// appear.
type WaitForDeviceOpts struct {
	LocalDevicesOpts

	// Token is the value returned by a remote VolumeAttach call that the
	// client can use to block until a specific device has appeared in the
	// local devices list.
	Token string

	// Timeout is the maximum duration for which to wait for a device to
	// appear in the local devices list.
	Timeout time.Duration
}

// NewStorageExecutor is a function that constructs a new StorageExecutors.
type NewStorageExecutor func() StorageExecutor

// StorageExecutor is the part of a storage driver that is downloaded at
// runtime by the libStorage client.
type StorageExecutor interface {
	Driver
	StorageExecutorFunctions
}

// StorageExecutorFunctions is the collection of functions that are required of
// a StorageExecutor.
type StorageExecutorFunctions interface {

	// InstanceID returns the local system's InstanceID.
	InstanceID(
		ctx Context,
		opts Store) (*InstanceID, error)

	// NextDevice returns the next available device.
	NextDevice(
		ctx Context,
		opts Store) (string, error)

	// LocalDevices returns a map of the system's local devices.
	LocalDevices(
		ctx Context,
		opts *LocalDevicesOpts) (*LocalDevices, error)
}

// StorageExecutorWithSupported is an interface that executor implementations
// may use by defining the function "Supported(Context, Store) (bool, error)".
// This function indicates whether a storage platform is valid when executing
// the executor binary on a given client.
type StorageExecutorWithSupported interface {
	StorageExecutorFunctions

	// Supported returns a flag indicating whether or not the platform
	// implementing the executor is valid for the host on which the executor
	// resides.
	Supported(
		ctx Context,
		opts Store) (bool, error)
}

// StorageExecutorWithMount is an interface that executor implementations
// may use to become part of the mount workflow.
type StorageExecutorWithMount interface {

	// Mount mounts a device to a specified path.
	Mount(
		ctx Context,
		deviceName, mountPoint string,
		opts *DeviceMountOpts) error
}

// StorageExecutorWithMounts is an interface that executor implementations
// may use to become part of the mounts workflow.
type StorageExecutorWithMounts interface {

	// Mounts get a list of mount points.
	Mounts(
		ctx Context,
		opts Store) ([]*MountInfo, error)
}

// StorageExecutorWithUnmount is an interface that executor implementations
// may use to become part of unmount workflow.
type StorageExecutorWithUnmount interface {

	// Unmount unmounts the underlying device from the specified path.
	Unmount(
		ctx Context,
		mountPoint string,
		opts Store) error
}

// StorageExecutorWithVolumeCreate is an interface that executor implementations
// may use to become part of create workflow.
type StorageExecutorWithVolumeCreate interface {

	// VolumeCreate creates a new volume.
	VolumeCreate(
		ctx Context,
		name string,
		opts *VolumeCreateOpts) (*Volume, error)
}

// StorageExecutorWithVolumeRemove is an interface that executor implementations
// may use to become part of remove workflow.
type StorageExecutorWithVolumeRemove interface {

	// VolumeRemove removes a volume.
	VolumeRemove(
		ctx Context,
		volumeID string,
		opts *VolumeRemoveOpts) error
}

// StorageExecutorWithVolumeAttach will attach a volume based on volumeName to
// the instance of instanceID.
type StorageExecutorWithVolumeAttach interface {

	// VolumeAttach attaches a volume and provides a token clients can use
	// to validate that device has appeared locally.
	VolumeAttach(
		ctx Context,
		volumeID string,
		opts *VolumeAttachOpts) (*Volume, string, error)
}

// StorageExecutorWithVolumeDetach will detach a volume based on volumeName to
// the instance of instanceID.
type StorageExecutorWithVolumeDetach interface {

	// VolumeDetach detaches a volume.
	VolumeDetach(
		ctx Context,
		volumeID string,
		opts *VolumeDetachOpts) (*Volume, error)
}

// ProvidesStorageExecutorCLI is a type that provides the StorageExecutorCLI.
type ProvidesStorageExecutorCLI interface {

	// XCLI returns the StorageExecutorCLI.
	XCLI() StorageExecutorCLI
}

// LSXVolumeAttachResult is the object type used to marshal the possible tuple
// returned from the VolumeAttach call.
type LSXVolumeAttachResult struct {
	Volume *Volume `json:"volume,omitempty"`
	Token  string  `json:"token,omitempty"`
}

// StorageExecutorCLI provides a way to interact with the CLI tool built with
// the driver implementations of the StorageExecutor interface.
type StorageExecutorCLI interface {
	StorageExecutorFunctions
	StorageExecutorWithMount
	StorageExecutorWithMounts
	StorageExecutorWithUnmount

	// WaitForDevice blocks until the provided attach token appears in the
	// map returned from LocalDevices or until the timeout expires, whichever
	// occurs first.
	//
	// The return value is a boolean flag indicating whether or not a match was
	// discovered as well as the result of the last LocalDevices call before a
	// match is discovered or the timeout expires.
	WaitForDevice(
		ctx Context,
		opts *WaitForDeviceOpts) (bool, *LocalDevices, error)

	// Supported returns a flag indicating whether the executor supports
	// specific functions for a storage platform on the current host.
	Supported(
		ctx Context,
		opts Store) (LSXSupportedOp, error)

	// LSXVolumeCreate creates a new volume.
	LSXVolumeCreate(
		ctx Context,
		name string,
		opts *VolumeCreateOpts) (*Volume, error)

	// LSXVolumeRemove removes a volume.
	LSXVolumeRemove(
		ctx Context,
		volumeID string,
		opts *VolumeRemoveOpts) error

	// LSXVolumeAttach attaches a volume and provides a token clients can use
	// to validate that device has appeared locally.
	LSXVolumeAttach(
		ctx Context,
		volumeID string,
		opts *VolumeAttachOpts) (*Volume, string, error)

	// LSXVolumeDetach detaches a volume.
	LSXVolumeDetach(
		ctx Context,
		volumeID string,
		opts *VolumeDetachOpts) (*Volume, error)
}

// LSXSupportedOp is a bit for the mask returned from an executor's Supported
// function.
type LSXSupportedOp int

const (
	// LSXSOpInstanceID indicates an executor supports "InstanceID".
	// "InstanceID" operation.
	LSXSOpInstanceID LSXSupportedOp = 1 << iota // 1

	// LSXSOpNextDevice indicates an executor supports "NextDevice".
	LSXSOpNextDevice

	// LSXSOpLocalDevices indicates an executor supports "LocalDevices".
	LSXSOpLocalDevices

	// LSXSOpWaitForDevice indicates an executor supports "WaitForDevice".
	LSXSOpWaitForDevice

	// LSXSOpMount indicates an executor supports "Mount".
	LSXSOpMount

	// LSXSOpUmount indicates an executor supports "Umount".
	LSXSOpUmount

	// LSXSOpMounts indicates an executor supports "Mounts".
	LSXSOpMounts

	// LSXSOpVolumeCreate indicates an executor supports "VolumeCreate".
	LSXSOpVolumeCreate

	// LSXSOpVolumeRemove indicates an executor supports "VolumeRemove".
	LSXSOpVolumeRemove

	// LSXSOpVolumeAttach indicates an executor supports "VolumeAttach".
	LSXSOpVolumeAttach

	// LSXSOpVolumeDetach indicates an executor supports "VolumeDetach".
	LSXSOpVolumeDetach
)

const (
	// LSXSOpNone indicates the executor is not supported for the platform.
	LSXSOpNone LSXSupportedOp = 0

	// LSXOpClassic indicates the classic executor support.
	LSXOpClassic = LSXSOpInstanceID |
		LSXSOpLocalDevices |
		LSXSOpNextDevice |
		LSXSOpWaitForDevice
)

// InstanceID returns a flag that indicates whether the LSXSOpInstanceID bit
// is set.
func (v LSXSupportedOp) InstanceID() bool {
	return v.bitSet(LSXSOpInstanceID)
}

// NextDevice returns a flag that indicates whether the LSXSOpNextDevice bit
// is set.
func (v LSXSupportedOp) NextDevice() bool {
	return v.bitSet(LSXSOpNextDevice)
}

// LocalDevices returns a flag that indicates whether the LSXSOpLocalDevices
// bit is set.
func (v LSXSupportedOp) LocalDevices() bool {
	return v.bitSet(LSXSOpLocalDevices)
}

// WaitForDevice returns a flag that indicates whether the LSXSOpWaitForDevice
// bit is set.
func (v LSXSupportedOp) WaitForDevice() bool {
	return v.bitSet(LSXSOpWaitForDevice)
}

// Mount returns a flag that indicates whether the LSXSOpMount bit
// is set.
func (v LSXSupportedOp) Mount() bool {
	return v.bitSet(LSXSOpMount)
}

// Umount returns a flag that indicates whether the LSXSOpUmount bit
// is set.
func (v LSXSupportedOp) Umount() bool {
	return v.bitSet(LSXSOpUmount)
}

// Mounts returns a flag that indicates whether the LSXSOpMounts bit
// is set.
func (v LSXSupportedOp) Mounts() bool {
	return v.bitSet(LSXSOpMounts)
}

// VolumeCreate returns a flag that indicates whether the LSXSOpVolumeCreate bit
// is set.
func (v LSXSupportedOp) VolumeCreate() bool {
	return v.bitSet(LSXSOpVolumeCreate)
}

// VolumeRemove returns a flag that indicates whether the LSXSOpVolumeRemove bit
// is set.
func (v LSXSupportedOp) VolumeRemove() bool {
	return v.bitSet(LSXSOpVolumeRemove)
}

// VolumeAttach returns a flag that indicates whether the LSXSOpVolumeAttach bit
// is set.
func (v LSXSupportedOp) VolumeAttach() bool {
	return v.bitSet(LSXSOpVolumeAttach)
}

// VolumeDetach returns a flag that indicates whether the LSXSOpVolumeDetach bit
// is set.
func (v LSXSupportedOp) VolumeDetach() bool {
	return v.bitSet(LSXSOpVolumeDetach)
}

func (v LSXSupportedOp) bitSet(b LSXSupportedOp) bool {
	return v&b == b
}
