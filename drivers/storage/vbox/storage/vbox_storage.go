// +build !libstorage_storage_driver libstorage_storage_driver_vbox

package storage

import (
	"path/filepath"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"

	"github.com/akutz/goof"
	vboxw "github.com/appropriate/go-virtualboxclient/vboxwebsrv"
	vboxc "github.com/appropriate/go-virtualboxclient/virtualboxclient"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/drivers/storage/vbox"
)

// Driver represents a vbox driver implementation of StorageDriver
type driver struct {
	sync.Mutex
	config gofig.Config
}

func init() {
	registry.RegisterStorageDriver(vbox.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

// Name returns the name of the driver
func (d *driver) Name() string {
	return vbox.Name
}

// Init initializes the driver.
func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	fields := map[string]interface{}{
		"provider":        vbox.Name,
		"moduleName":      vbox.Name,
		"endpoint":        d.endpoint(),
		"userName":        d.username(),
		"tls":             d.tls(),
		"volumePath":      d.volumePath(),
		"controllerName":  d.controllerName(),
		"machineNameOrId": d.machineNameID(""),
	}
	ctx.WithFields(fields).Info("storage driver initialized")
	return nil
}

func (d *driver) Login(ctx types.Context) (interface{}, error) {
	svc := vboxc.New(
		d.username(),
		d.password(),
		d.endpoint(),
		d.tls(),
		d.controllerName())
	if err := svc.Logon(); err != nil {
		return err, goof.WithError("error logging into vbox", err)
	}
	return svc, nil
}

func mustSession(ctx types.Context) *vboxc.VirtualBox {
	return context.MustSession(ctx).(*vboxc.VirtualBox)
}

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context, opts types.Store) (*types.LocalDevices, error) {
	if ld, ok := context.LocalDevices(ctx); ok {
		return ld, nil
	}
	return nil, goof.New("missing local devices")
}

// NextDevice returns the next available device (not implemented).
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", nil
}

// Type returns the type of storage a driver provides
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

// NextDeviceInfo returns the information about the driver's next available
// device workflow.
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	if iid.ID != "" {
		return &types.Instance{
			Name:         iid.ID,
			InstanceID:   iid,
			ProviderName: iid.Driver,
		}, nil
	}

	var macAddrs []string
	if err := iid.UnmarshalMetadata(&macAddrs); err != nil {
		return nil, err
	}

	svcObj, err := d.Login(ctx)
	if err != nil {
		return nil, err
	}
	svc, ok := svcObj.(*vboxc.VirtualBox)
	if !ok {
		return nil, goof.New("invalid service type")
	}
	ctx = context.WithValue(ctx, context.SessionKey, svc)

	var m *vboxc.Machine
	if m, err = d.findMachineByMacAddrs(ctx, macAddrs); err != nil {
		nameOrID := d.config.GetString("virtualbox.localMachineNameOrId")
		if m, err = d.findMachineByNameOrID(ctx, nameOrID); err != nil {
			return nil, goof.WithFieldsE(
				log.Fields{
					"macAddres":                       macAddrs,
					"virtualbox.localMachineNameOrId": nameOrID,
				}, "failed to find local machine", err)
		}
	}

	if m == nil {
		return nil, goof.New("machine not found")
	}

	return &types.Instance{
		Name: m.ID,
		InstanceID: &types.InstanceID{
			ID:     m.ID,
			Driver: vbox.Name,
		},
		ProviderName: vbox.Name,
	}, nil
}

// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) getVolumeMapping(ctx types.Context) ([]*types.Volume, error) {

	var (
		err         error
		mapDiskByID map[string]string
		mas         []*vboxw.IMediumAttachment
		m           *vboxc.Machine
		iid         = context.MustInstanceID(ctx)
	)

	m, err = d.findMachineByInstanceID(ctx, iid)
	if err != nil {
		ctx.WithField(
			"iid", iid.ID).WithError(
			err).Error("error finding machine by instance ID")
		return nil, err
	}

	if err := m.Refresh(); err != nil {
		ctx.WithError(err).Error("error refreshing machine ref")
		return nil, err
	}
	defer m.Release()

	ld, ok := context.LocalDevices(ctx)
	if !ok {
		return nil, goof.New("missing local devices")
	}
	mapDiskByID = ld.DeviceMap

	mas, err = m.GetMediumAttachments()
	if err != nil {
		ctx.WithError(err).Error("error getting medium attachments")
		return nil, err
	}

	processMediumAttachment := func(
		ctx types.Context,
		svc *vboxc.VirtualBox,
		ma *vboxw.IMediumAttachment) (*types.Volume, bool, error) {

		ctx.WithFields(log.Fields{
			"moref":      ma.Medium,
			"controller": ma.Controller,
			"device":     ma.Device,
			"port":       ma.Port,
			"type":       ma.Type_,
		}).Debug("processing medium")

		medium := svc.NewMedium(ma.Medium)
		defer medium.Release()

		mid, err := medium.GetID()
		if err != nil {
			ctx.WithError(err).Error("error getting medium ID")
			return nil, false, err
		}
		ctx.WithField("id", mid).Debug("processing medium")

		smid := strings.Split(mid, "-")
		if len(smid) == 0 {
			return nil, true, nil
		}

		location, err := medium.GetLocation()
		if err != nil {
			ctx.WithError(err).Error("error getting medium location")
			return nil, false, err
		}
		ctx.WithField("location", location).Debug("processing medium")

		var bdn string
		var ok bool
		if bdn, ok = mapDiskByID[smid[0]]; !ok {
			return nil, true, nil
		}

		return &types.Volume{
			Name:   bdn,
			ID:     mid,
			Status: location,
		}, false, nil
	}

	var (
		svc          = mustSession(ctx)
		blockDevices []*types.Volume
	)

	for _, ma := range mas {
		vol, cont, err := processMediumAttachment(ctx, svc, ma)
		if err != nil {
			return nil, err
		}
		if cont {
			continue
		}
		blockDevices = append(blockDevices, vol)
	}
	return blockDevices, nil
}

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	d.Lock()
	defer d.Unlock()

	if opts.Size == nil {
		return nil, goof.New("missing volume size")
	}

	fields := map[string]interface{}{
		"provider":   vbox.Name,
		"volumeName": volumeName,
		"size":       *opts.Size,
	}

	size := *opts.Size * 1024 * 1024 * 1024

	vol, err := d.getVolume(ctx, "", volumeName, types.VolAttFalse)
	if err != nil {
		return nil, err
	}

	if vol != nil {
		return nil, goof.New("volume already exists")
	}

	med, err := d.createVolume(ctx, volumeName, size)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error creating new volume", err)
	}

	var iops int64
	if opts.IOPS != nil {
		iops = *opts.IOPS
	}

	newVol := &types.Volume{
		ID:   med.ID,
		Name: med.Name,
		Size: med.LogicalSize / 1024 / 1024 / 1024,
		IOPS: iops,
		Type: string(med.DeviceType),
	}

	return newVol, nil
}

// VolumeCreateFromSnapshot (not implemented).
func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

// VolumeCopy copies an existing volume (not implemented)
func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

// VolumeSnapshot snapshots a volume (not implemented)
func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	d.Lock()
	defer d.Unlock()

	fields := map[string]interface{}{
		"provider": vbox.Name,
		"volumeID": volumeID,
	}

	err := mustSession(ctx).RemoveMedium(volumeID)
	if err != nil {
		return goof.WithFieldsE(fields, "error deleting volume", err)
	}

	return nil
}

// VolumeAttach attaches a volume.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	if volumeID == "" {
		return nil, "", goof.New("missing volume id")
	}

	// review volume with attachments to any host
	volumes, err := d.getVolume(ctx, volumeID, "", types.VolAttReq)
	if err != nil {
		return nil, "", err
	}

	if len(volumes) == 0 {
		return nil, "", goof.New("no volume found")
	}

	if len(volumes[0].Attachments) > 0 && !opts.Force {
		return nil, "", goof.New("volume already attached to a host")
	}
	if opts.Force {
		if _, err := d.VolumeDetach(ctx, volumeID, nil); err != nil {
			return nil, "", err
		}
	}

	err = d.attachVolume(ctx, volumeID, "")
	if err != nil {
		return nil, "", goof.WithFieldsE(
			log.Fields{
				"provider": vbox.Name,
				"volumeID": volumeID},
			"error attaching volume",
			err,
		)
	}

	volumes, err = d.getVolume(ctx, volumeID, "", types.VolAttReqTrue)
	if err != nil {
		return nil, "", err
	}

	if len(volumes) == 0 {
		return nil, "", err
	}

	svid := strings.Split(volumes[0].ID, "-")

	return volumes[0], svid[0], nil
}

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	if volumeID == "" {
		return nil, goof.New("missing volume id")
	}

	volumes, err := d.getVolume(ctx, volumeID, "", types.VolAttFalse)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("no volume returned")
	}

	// TODO: Check if volumes[[0].Attachments > 0?

	if err := d.detachVolume(ctx, volumeID, ""); err != nil {
		return nil, goof.WithFieldsE(
			log.Fields{
				"provier":  vbox.Name,
				"volumeID": volumeID}, "error detaching volume", err)
	}

	ctx.Info("detached volume", volumeID)
	return d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{
			Attachments: types.VolAttReqTrue})
}

func (d *driver) VolumeDetachAll(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {
	return nil
}

func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	return nil
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	vols, err := d.getVolume(ctx, "", "", opts.Attachments)
	if err != nil {
		return nil, err
	}
	return vols, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	vols, err := d.getVolume(ctx, volumeID, "", opts.Attachments)
	if err != nil {
		return nil, err
	}
	if len(vols) == 0 {
		return nil, goof.New("no volumes returned")
	}
	return vols[0], nil
}

// getVolume searches and returns a volume matching criteria
func (d *driver) getVolume(
	ctx types.Context,
	volumeID string, volumeName string,
	attachments types.VolumeAttachmentsTypes) ([]*types.Volume, error) {

	volumes, err := mustSession(ctx).GetMedium(volumeID, volumeName)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, nil
	}

	var mapDN map[string]string
	if attachments.Devices() {
		volumeMapping, err := d.getVolumeMapping(ctx)
		if err != nil {
			return nil, err
		}

		mapDN = make(map[string]string)
		for _, vm := range volumeMapping {
			mapDN[vm.ID] = vm.Name
		}
	}

	var volumesSD []*types.Volume

	for _, v := range volumes {
		volumeSD := &types.Volume{
			Name:   v.Name,
			ID:     v.ID,
			Size:   int64(v.LogicalSize / 1024 / 1024 / 1024),
			Status: v.Location,
			Type:   string(v.DeviceType),
		}

		if attachments.Requested() {
			var attachmentsSD []*types.VolumeAttachment
			for _, mid := range v.MachineIDs {
				attachmentSD := &types.VolumeAttachment{
					VolumeID: v.ID,
					InstanceID: &types.InstanceID{
						ID:     mid,
						Driver: vbox.Name,
					},
				}
				if attachments.Devices() && mapDN != nil {
					dn, _ := mapDN[v.ID]
					attachmentSD.DeviceName = dn
				}
				attachmentsSD = append(attachmentsSD, attachmentSD)
			}
			volumeSD.Attachments = attachmentsSD
		}

		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

func (d *driver) findMachineByInstanceID(
	ctx types.Context,
	iid *types.InstanceID) (*vboxc.Machine, error) {
	return d.findMachineByNameOrID(ctx, iid.ID)
}

func (d *driver) findMachineByNameOrID(
	ctx types.Context,
	nameOrID string) (*vboxc.Machine, error) {

	ctx.WithField("nameOrID", nameOrID).Debug("finding local machine")

	m, err := mustSession(ctx).FindMachine(nameOrID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, goof.New("could not find machine")
	}

	if id, err := m.GetID(); err == nil {
		m.ID = id
	} else {
		return nil, err
	}

	if name, err := m.GetName(); err == nil {
		m.Name = name
	} else {
		return nil, err
	}

	return m, nil
}

func (d *driver) findMachineByMacAddrs(
	ctx types.Context,
	macAddrs []string) (*vboxc.Machine, error) {
	ctx.WithField("macAddrs", macAddrs).Debug("finding local machine")

	macMap := make(map[string]bool)
	for _, mac := range macAddrs {
		macUp := mac
		macMap[macUp] = true
	}

	svc := mustSession(ctx)

	machines, err := svc.GetMachines()
	if err != nil {
		return nil, err
	}

	sp, err := svc.GetSystemProperties()
	if err != nil {
		return nil, err
	}
	defer sp.Release()

	for _, m := range machines {
		defer m.Release()
		chipset, err := m.GetChipsetType()
		if err != nil {
			return nil, err
		}

		mna, err := sp.GetMaxNetworkAdapters(chipset)
		if err != nil {
			return nil, err
		}

		for i := uint32(0); i < mna; i++ {
			na, err := m.GetNetworkAdapter(i)
			if err != nil {
				return nil, err
			}

			mac, err := na.GetMACAddress()
			if err != nil {
				return nil, err
			}

			if _, ok := macMap[mac]; ok {
				id, err := m.GetID()
				if err != nil {
					return nil, err
				}
				m.ID = id

				name, err := m.GetName()
				if err != nil {
					return nil, err
				}
				m.Name = name

				return m, nil
			}
		}
	}

	return nil, goof.New("unable to find machine")
}

func (d *driver) createVolume(
	ctx types.Context, name string, size int64) (*vboxc.Medium, error) {

	if name == "" {
		return nil, goof.New("name is empty")
	}
	path := filepath.Join(d.volumePath(), name)
	ctx.WithField("path", path).Debug("creating vmdk")
	return mustSession(ctx).CreateMedium("vmdk", path, size)
}

func (d *driver) attachVolume(
	ctx types.Context, volumeID, volumeName string) error {

	iid := context.MustInstanceID(ctx)

	m, err := d.findMachineByInstanceID(ctx, iid)
	if err != nil {
		return err
	}

	if err := m.Refresh(); err != nil {
		return err
	}
	defer m.Release()

	medium, err := mustSession(ctx).GetMedium(volumeID, volumeName)
	if err != nil {
		return err
	}

	if len(medium) == 0 {
		return goof.New("no volume returned")
	}
	if len(medium) > 1 {
		return goof.New("too many volumes returned")
	}

	if err := m.AttachDevice(medium[0]); err != nil {
		return err
	}

	return nil
}

func (d *driver) detachVolume(
	ctx types.Context, volumeID, volumeName string) error {

	iid := context.MustInstanceID(ctx)

	m, err := d.findMachineByInstanceID(ctx, iid)
	if err != nil {
		return err
	}

	if err := m.Refresh(); err != nil {
		return err
	}
	defer m.Release()

	media, err := mustSession(ctx).GetMedium(volumeID, volumeName)
	if err != nil {
		return err
	}

	if len(media) == 0 {
		return goof.New("no volume returned")
	}
	if len(media) > 1 {
		return goof.New("too many volumes returned")
	}

	if err := media[0].DetachMachines(); err != nil {
		return err
	}

	return nil
}

func (d *driver) username() string {
	return d.config.GetString("virtualbox.username")
}

func (d *driver) password() string {
	return d.config.GetString("virtualbox.password")
}

func (d *driver) endpoint() string {
	return d.config.GetString("virtualbox.endpoint")
}

func (d *driver) volumePath() string {
	return d.config.GetString("virtualbox.volumePath")
}

func (d *driver) controllerName() string {
	return d.config.GetString("virtualbox.controllerName")
}

func (d *driver) tls() bool {
	return d.config.GetBool("virtualbox.tls")
}

func (d *driver) machineNameID(nameOrID string) string {
	if nameOrID != "" {
		return nameOrID
	}
	return d.config.GetString("virtualbox.localMachineNameOrId")
}
