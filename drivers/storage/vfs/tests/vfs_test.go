package vfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/server"
	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"

	// load the vfs driver packages

	"github.com/emccode/libstorage/drivers/storage/vfs"
	_ "github.com/emccode/libstorage/drivers/storage/vfs/client"
	_ "github.com/emccode/libstorage/drivers/storage/vfs/storage"
)

func TestMain(m *testing.M) {
	server.CloseOnAbort()
	ec := m.Run()
	removeTestDirs()
	os.Exit(ec)
}

func TestClient(t *testing.T) {
	apitests.Run(t, vfs.Name, newTestConfigYAML(t),
		func(config gofig.Config, client types.Client, t *testing.T) {
			ctx := context.Background()
			iid, err := client.Executor().InstanceID(
				ctx.WithValue(context.ServiceKey, vfs.Name),
				utils.NewStore())
			assert.NoError(t, err)
			assert.NotNil(t, iid)
		})
}

func TestRoot(t *testing.T) {
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), apitests.TestRoot)
}

var testServicesFunc = func(
	config gofig.Config, client types.Client, t *testing.T) {

	reply, err := client.API().Services(nil)
	assert.NoError(t, err)
	assert.Equal(t, len(reply), 1)

	_, ok := reply[vfs.Name]
	assert.True(t, ok)
}

func TestServices(t *testing.T) {
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), testServicesFunc)
}

func TestServicesWithControllerClient(t *testing.T) {
	apitests.RunWithClientType(
		t, types.ControllerClient, vfs.Name,
		newTestConfigYAML(t), testServicesFunc)
}

func TestServiceInpspect(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {

		reply, err := client.API().ServiceInspect(nil, vfs.Name)
		assert.NoError(t, err)
		assert.Equal(t, vfs.Name, reply.Name)
		assert.Equal(t, vfs.Name, reply.Driver.Name)
		assert.True(t, reply.Driver.NextDevice.Ignore)
	}
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf)
}

func TestExecutors(t *testing.T) {
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), apitests.TestExecutors)
}

func TestExecutorsWithControllerClient(t *testing.T) {
	apitests.RunWithClientType(
		t, types.ControllerClient, vfs.Name, newTestConfigYAML(t),
		apitests.TestExecutorsWithControllerClient)
}

func TestExecutorHead(t *testing.T) {
	apitests.RunGroup(
		t, vfs.Name, newTestConfigYAML(t),
		apitests.TestHeadExecutorLinux,
		apitests.TestHeadExecutorDarwin)
	//apitests.TestHeadExecutorWindows)
}

func TestExecutorGet(t *testing.T) {
	apitests.RunGroup(
		t, vfs.Name, newTestConfigYAML(t),
		apitests.TestGetExecutorLinux,
		apitests.TestGetExecutorDarwin)
	//apitests.TestGetExecutorWindows)
}

func TestStorageDriverVolumes(t *testing.T) {
	apitests.Run(t, vfs.Name, newTestConfigYAML(t),
		func(config gofig.Config, client types.Client, t *testing.T) {

			vols, err := client.Storage().Volumes(
				context.Background().WithValue(
					context.ServiceKey, vfs.Name),
				&types.VolumesOpts{Attachments: true, Opts: utils.NewStore()})
			assert.NoError(t, err)
			assert.Len(t, vols, 2)
		})
}

func TestVolumes(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Volumes(nil, false)
		if err != nil {
			t.Fatal(err)
		}
		for volumeID, volume := range tc.volMap {
			assert.NotNil(t, reply["vfs"][volumeID])
			assert.EqualValues(t, volume, reply["vfs"][volumeID])
		}
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
	apitests.RunWithClientType(
		t, types.ControllerClient, vfs.Name, tc.configYAML, tf)
}

func TestVolumesWithAttachments(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Volumes(nil, true)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, reply["vfs"]["vfs-000"])
		assert.NotNil(t, reply["vfs"]["vfs-001"])
		assert.Nil(t, reply["vfs"]["vfs-002"])
		assert.EqualValues(t, tc.volMap["vfs-000"], reply["vfs"]["vfs-000"])
		assert.EqualValues(t, tc.volMap["vfs-001"], reply["vfs"]["vfs-001"])
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestVolumesWithAttachmentsWithControllerClient(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {

		_, err := client.API().Volumes(nil, true)
		assert.Error(t, err)
		assert.Equal(t, "batch processing error", err.Error())
	}

	apitests.RunWithClientType(
		t, types.ControllerClient, vfs.Name, tc.configYAML, tf)
}

func TestVolumesByService(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().VolumesByService(nil, "vfs", false)
		if err != nil {
			t.Fatal(err)
		}
		for volumeID, volume := range tc.volMap {
			assert.NotNil(t, reply[volumeID])
			assert.EqualValues(t, volume, reply[volumeID])
		}
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestVolumesByServiceWithAttachments(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().VolumesByService(nil, "vfs", true)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, reply["vfs-000"])
		assert.NotNil(t, reply["vfs-001"])
		assert.Nil(t, reply["vfs-002"])
		assert.EqualValues(t, tc.volMap["vfs-000"], reply["vfs-000"])
		assert.EqualValues(t, tc.volMap["vfs-001"], reply["vfs-001"])
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestVolumeInspect(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().VolumeInspect(nil, "vfs", "vfs-000", false)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, reply)
		assert.EqualValues(t, tc.volMap[reply.ID], reply)
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestVolumeInspectWithAttachments(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().VolumeInspect(nil, "vfs", "vfs-000", true)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, reply)
		assert.EqualValues(t, tc.volMap[reply.ID], reply)
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestSnapshots(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Snapshots(nil)
		if err != nil {
			t.Fatal(err)
		}
		for snapshotID, snapshot := range tc.snapMap {
			assert.NotNil(t, reply["vfs"][snapshotID])
			assert.EqualValues(t, snapshot, reply["vfs"][snapshotID])
		}
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestSnapshotsByService(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().SnapshotsByService(nil, "vfs")
		if err != nil {
			t.Fatal(err)
		}
		for snapshotID, snapshot := range tc.snapMap {
			assert.NotNil(t, reply[snapshotID])
			assert.EqualValues(t, snapshot, reply[snapshotID])
		}
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestVolumeCreate(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		volumeName := "Volume 003"
		availabilityZone := "US"
		iops := int64(1000)
		size := int64(10240)
		volType := "myType"

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		request := &types.VolumeCreateRequest{
			Name:             volumeName,
			AvailabilityZone: &availabilityZone,
			IOPS:             &iops,
			Size:             &size,
			Type:             &volType,
			Opts:             opts,
		}

		reply, err := client.API().VolumeCreate(nil, vfs.Name, request)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		assert.NotNil(t, reply)

		assertVolDir(t, config, reply.ID, true)
		assert.Equal(t, availabilityZone, reply.AvailabilityZone)
		assert.Equal(t, iops, reply.IOPS)
		assert.Equal(t, volumeName, reply.Name)
		assert.Equal(t, size, reply.Size)
		assert.Equal(t, volType, reply.Type)
		assert.Equal(t, "2", reply.Fields["priority"])
		assert.Equal(t, "root@example.com", reply.Fields["owner"])
	}

	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf)
}

func TestVolumeCopy(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		request := &types.VolumeCopyRequest{
			VolumeName: "Copy of Volume 000",
			Opts: map[string]interface{}{
				"priority": 7,
				"owner":    "user@example.com",
			},
		}

		reply, err := client.API().VolumeCopy(nil, vfs.Name, "vfs-000", request)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}

		assert.NotNil(t, reply)

		assertVolDir(t, config, reply.ID, true)
		assert.Equal(t, "vfs-003", reply.ID)
		assert.Equal(t, request.VolumeName, reply.Name)
		assert.Equal(t, "7", reply.Fields["priority"])
		assert.Equal(t, request.Opts["owner"], reply.Fields["owner"])
	}

	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf)
}

func TestVolumeRemove(t *testing.T) {

	tf1 := func(config gofig.Config, client types.Client, t *testing.T) {
		assertVolDir(t, config, "vfs-002", true)
		err := client.API().VolumeRemove(nil, vfs.Name, "vfs-002")
		assert.NoError(t, err)
		assertVolDir(t, config, "vfs-002", false)
	}

	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf1)

	tf2 := func(config gofig.Config, client types.Client, t *testing.T) {
		err := client.API().VolumeRemove(nil, vfs.Name, "vfs-002")
		assert.Error(t, err)
		httpErr := err.(goof.HTTPError)
		assert.Equal(t, "resource not found", httpErr.Error())
		assert.Equal(t, 404, httpErr.Status())
	}

	apitests.RunGroup(t, vfs.Name, newTestConfigYAML(t), tf1, tf2)
}

func TestVolumeSnapshot(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		volumeID := "vfs-000"
		snapshotName := "snapshot1"
		opts := map[string]interface{}{
			"priority": 2,
		}

		request := &types.VolumeSnapshotRequest{
			SnapshotName: snapshotName,
			Opts:         opts,
		}

		reply, err := client.API().VolumeSnapshot(
			nil, vfs.Name, volumeID, request)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		assert.Equal(t, snapshotName, reply.Name)
		assert.Equal(t, volumeID, reply.VolumeID)

		snap, err := client.API().SnapshotInspect(nil, vfs.Name, reply.ID)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		assert.Equal(t, snapshotName, snap.Name)
		assert.Equal(t, volumeID, snap.VolumeID)

		snapshots, err := client.API().SnapshotsByService(nil, vfs.Name)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		assert.EqualValues(t, 10, len(snapshots))
	}
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf)
}

func TestVolumeCreateFromSnapshot(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {

		ogVol, err := client.API().VolumeInspect(nil, "vfs", "vfs-000", true)
		assert.NoError(t, err)

		volumeName := "Volume 003"
		size := int64(40960)

		request := &types.VolumeCreateRequest{
			Name: volumeName,
			Size: &size,
			Opts: map[string]interface{}{
				"priority": 4,
				"owner":    "user@example.com",
			},
		}

		reply, err := client.API().VolumeCreateFromSnapshot(
			nil, vfs.Name, "vfs-000-002", request)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}

		assert.NotNil(t, reply)

		assertVolDir(t, config, reply.ID, true)
		assert.Equal(t, ogVol.AvailabilityZone, reply.AvailabilityZone)
		assert.Equal(t, ogVol.IOPS, reply.IOPS)
		assert.Equal(t, volumeName, reply.Name)
		assert.Equal(t, size, reply.Size)
		assert.Equal(t, ogVol.Type, reply.Type)
		assert.Equal(t, "4", reply.Fields["priority"])
		assert.Equal(t, request.Opts["owner"], reply.Fields["owner"])

	}
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf)
}

func TestVolumeAttach(t *testing.T) {

	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {

		nextDevice, err := client.Executor().NextDevice(
			context.Background().WithValue(context.ServiceKey, vfs.Name),
			utils.NewStore())
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}

		req := &types.VolumeAttachRequest{
			NextDeviceName: &nextDevice,
		}

		volID := "vfs-002"

		reply, attTokn, err := client.API().VolumeAttach(
			nil, vfs.Name, volID, req)
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}

		if reply == nil {
			assert.NotNil(t, reply)
			t.FailNow()
		}

		assert.Equal(t, volID, attTokn)
		assert.Equal(t, volID, reply.ID)
		assert.Equal(t, req.NextDeviceName, reply.Attachments[0].DeviceName)

	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestVolumeAttachWithControllerClient(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {

		_, err := client.Executor().NextDevice(
			context.Background().WithValue(context.ServiceKey, vfs.Name),
			utils.NewStore())
		assert.Error(t, err)
		assert.Equal(t, "unsupported op for client type", err.Error())

		_, _, err = client.API().VolumeAttach(
			nil, vfs.Name, "vfs-002", &types.VolumeAttachRequest{})
		assert.Error(t, err)
		assert.Equal(t, "unsupported op for client type", err.Error())
	}

	apitests.RunWithClientType(
		t, types.ControllerClient, vfs.Name, newTestConfigYAML(t), tf)
}

func TestVolumeDetach(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		request := &types.VolumeDetachRequest{}
		reply, err := client.API().VolumeDetach(
			nil, vfs.Name, "vfs-001", request)
		assert.NoError(t, err)
		assert.Equal(t, "vfs-001", reply.ID)
		assert.Equal(t, 0, len(reply.Attachments))

		reply, err = client.API().VolumeInspect(
			nil, vfs.Name, "vfs-001", false)
		assert.NoError(t, err)
		assert.Equal(t, "vfs-001", reply.ID)
		assert.Equal(t, 0, len(reply.Attachments))
	}
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf)
}

func TestVolumeDetachWithControllerClient(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		_, err := client.API().VolumeDetach(
			nil, vfs.Name, "vfs-001", &types.VolumeDetachRequest{})
		assert.Error(t, err)
		assert.Equal(t, "unsupported op for client type", err.Error())
	}
	apitests.RunWithClientType(
		t, types.ControllerClient, vfs.Name, newTestConfigYAML(t), tf)
}

func TestVolumeDetachAllForService(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		request := &types.VolumeDetachRequest{}
		reply, err := client.API().VolumeDetachAllForService(
			nil, vfs.Name, request)
		assert.NoError(t, err)
		for _, v := range tc.volMap {
			v.Attachments = nil
		}
		assert.Equal(t, 3, len(reply))
		assert.EqualValues(t, tc.volMap, reply)

		reply, err = client.API().VolumesByService(
			nil, vfs.Name, true)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(reply))

		reply, err = client.API().VolumesByService(
			nil, vfs.Name, false)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(reply))
		assert.EqualValues(t, tc.volMap, reply)
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestVolumeDetachAllForServiceWithControllerClient(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		_, err := client.API().VolumeDetachAllForService(
			nil, vfs.Name, &types.VolumeDetachRequest{})
		assert.Error(t, err)
		assert.Equal(t, "unsupported op for client type", err.Error())
	}
	apitests.RunWithClientType(
		t, types.ControllerClient, vfs.Name, newTestConfigYAML(t), tf)
}

func TestVolumeDetachAll(t *testing.T) {
	tc := newTestConfig(t)
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		request := &types.VolumeDetachRequest{}
		reply, err := client.API().VolumeDetachAll(
			nil, request)
		assert.NoError(t, err)
		for _, v := range tc.volMap {
			v.Attachments = nil
		}
		assert.Equal(t, 1, len(reply))
		assert.Equal(t, 3, len(reply[vfs.Name]))
		assert.EqualValues(t, tc.volMap, reply[vfs.Name])

		reply, err = client.API().Volumes(nil, true)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(reply))
		assert.Equal(t, 0, len(reply[vfs.Name]))

		reply, err = client.API().Volumes(nil, false)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(reply))
		assert.Equal(t, 3, len(reply[vfs.Name]))
	}
	apitests.Run(t, vfs.Name, tc.configYAML, tf)
}

func TestVolumeDetachAllWithControllerClient(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		_, err := client.API().VolumeDetachAll(
			nil, &types.VolumeDetachRequest{})
		assert.Error(t, err)
		assert.Equal(t, "unsupported op for client type", err.Error())
	}
	apitests.RunWithClientType(
		t, types.ControllerClient, vfs.Name, newTestConfigYAML(t), tf)
}

func TestSnapshotCopy(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		snapshotName := "Snapshot from vfs-000-000"

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		request := &types.SnapshotCopyRequest{
			SnapshotName: snapshotName,
			Opts:         opts,
		}

		reply, err := client.API().SnapshotCopy(
			nil, vfs.Name, "vfs-000-000", request)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)

		assert.Equal(t, snapshotName, reply.Name)
		assert.Equal(t, "vfs-000", reply.VolumeID)
		assert.Equal(t, "2", reply.Fields["priority"])
		assert.Equal(t, "root@example.com", reply.Fields["owner"])

	}
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf)
}

func TestSnapshotRemove(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().SnapshotInspect(nil, "vfs", "vfs-000-002")
		assert.NoError(t, err)
		assert.NotNil(t, reply)
		assert.Equal(t, "vfs-000-002", reply.ID)

		err = client.API().SnapshotRemove(nil, "vfs", reply.ID)
		assert.NoError(t, err)

		reply, err = client.API().SnapshotInspect(nil, "vfs", "vfs-000-002")
		assert.Error(t, err)
		assert.Nil(t, reply)

	}
	apitests.Run(t, vfs.Name, newTestConfigYAML(t), tf)
}

func TestInstanceID(t *testing.T) {
	iid, err := instanceID()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	apitests.Run(
		t, vfs.Name, newTestConfigYAML(t),
		(&apitests.InstanceIDTest{
			Driver:   vfs.Name,
			Expected: iid,
		}).Test)
}

func TestInstance(t *testing.T) {
	iid, err := instanceID()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	apitests.Run(
		t, vfs.Name, nil,
		(&apitests.InstanceTest{
			Driver: vfs.Name,
			Expected: &types.Instance{
				InstanceID: iid,
				Name:       "vfsInstance",
			},
		}).Test)
}

func TestNextDevice(t *testing.T) {
	tc := newTestConfig(t)
	apitests.RunGroup(
		t, vfs.Name, tc.configYAML,
		(&apitests.NextDeviceTest{
			Driver:   vfs.Name,
			Expected: path.Join(tc.vfsRoot, "dev", "xvdc"),
		}).Test)
}

func TestLocalDevices(t *testing.T) {
	tc := newTestConfig(t)
	apitests.RunGroup(
		t, vfs.Name, tc.configYAML,
		(&apitests.LocalDevicesTest{
			Driver: vfs.Name,
			Expected: &types.LocalDevices{
				Driver:    vfs.Name,
				DeviceMap: tc.devMap,
			},
		}).Test)
}

func removeTestDirs() {
	testDirsLock.RLock()
	defer testDirsLock.RUnlock()
	for _, d := range testDirs {
		os.RemoveAll(d)
		log.WithField("path", d).Debug("removed test dir")
	}
}

func instanceID() (*types.InstanceID, error) {
	hostName, err := utils.HostName()
	if err != nil {
		return nil, err
	}
	return &types.InstanceID{ID: hostName, Driver: vfs.Name}, nil
}

func assertVolDir(
	t *testing.T, config gofig.Config, volumeID string, exists bool) {
	volDir := path.Join(vfs.VolumesDirPath(config), volumeID)
	assert.Equal(t, exists, gotil.FileExists(volDir))
}

func assertSnapDir(
	t *testing.T, config gofig.Config, snapshotID string, exists bool) {
	snapDir := path.Join(vfs.SnapshotsDirPath(config), snapshotID)
	assert.Equal(t, exists, gotil.FileExists(snapDir))
}

var (
	testDirs     []string
	testDirsLock = &sync.RWMutex{}
)

type testConfig struct {
	vfsRoot    string
	configYAML []byte
	devMap     map[string]string
	volMap     map[string]*types.Volume
	snapMap    map[string]*types.Snapshot
}

func newTestConfigYAML(t *testing.T) []byte {
	tc := newTestConfig(t)
	return tc.configYAML
}

func newTestConfig(t *testing.T) *testConfig {

	var (
		err        error
		hostName   string
		devMapFile *os.File
		tc         = &testConfig{
			devMap:  map[string]string{},
			volMap:  map[string]*types.Volume{},
			snapMap: map[string]*types.Snapshot{},
		}
	)

	if hostName, err = utils.HostName(); err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	if tc.vfsRoot, err = ioutil.TempDir("", ""); err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	t.Logf("created temp vfs root dir: %s", tc.vfsRoot)

	func() {
		testDirsLock.Lock()
		defer testDirsLock.Unlock()
		testDirs = append(testDirs, tc.vfsRoot)
	}()

	vd := path.Join(tc.vfsRoot, "vol")
	if err = os.MkdirAll(vd, 0755); err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	t.Logf("created temp vfs vol dir: %s", vd)
	sd := path.Join(tc.vfsRoot, "snap")
	if err := os.MkdirAll(sd, 0755); err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	t.Logf("created temp vfs snap dir: %s", sd)

	tc.devMap["vfs-000"] = path.Join(tc.vfsRoot, "dev", "xvda")
	tc.devMap["vfs-001"] = path.Join(tc.vfsRoot, "dev", "xvdb")
	devMapFile, err = os.Open(path.Join(tc.vfsRoot, "dev.json"))
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	defer devMapFile.Close()

	enc := json.NewEncoder(devMapFile)
	if err = enc.Encode(tc.devMap); err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	t.Logf("created temp vfs dev file: %s", devMapFile.Name())

	for x := 0; x < 3; x++ {
		var vj []byte
		if x < 2 {
			vj = []byte(fmt.Sprintf(volJSON, x, hostName))
		} else {
			vj = []byte(fmt.Sprintf(volNoAttachJSON, x, hostName))
		}
		v := &types.Volume{}
		if err := json.Unmarshal(vj, v); err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		tc.volMap[v.ID] = v
		vjp := path.Join(vd, fmt.Sprintf("%s.json", v.ID))
		os.MkdirAll(path.Join(vd, v.ID), 0755)
		if err := ioutil.WriteFile(vjp, vj, 0644); err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}

		for y := 0; y < 3; y++ {
			sj := []byte(fmt.Sprintf(snapJSON, x, y, time.Now().Unix()))
			s := &types.Snapshot{}
			if err := json.Unmarshal(sj, s); err != nil {
				assert.NoError(t, err)
				t.FailNow()
			}
			tc.snapMap[s.ID] = s
			sjp := path.Join(sd, fmt.Sprintf("vfs-%03d-%03d.json", x, y))
			if err := ioutil.WriteFile(sjp, sj, 0644); err != nil {
				assert.NoError(t, err)
				t.FailNow()
			}
		}
	}

	tc.configYAML = []byte(fmt.Sprintf(configYAML, tc.vfsRoot))
	return tc
}

const configYAML = "vfs:\n  root: %s"

const volJSON = `{
    "availabilityZone": "US",
    "iops":             1000,
    "name":             "Volume %03[1]d",
    "size":             10240,
    "id":               "vfs-%03[1]d",
    "type":             "myType",
    "attachments": [{
        "volumeID":     "%03[1]d",
        "instanceID":   {
            "id":       "%[2]s"
        },
        "status":       "attached"
    }],
    "fields": {
        "owner":        "root@example.com",
        "priority":     "2"
    }
}`

const volNoAttachJSON = `{
    "availabilityZone": "US",
    "iops":             1000,
    "name":             "Volume %03[1]d",
    "size":             10240,
    "id":               "vfs-%03[1]d",
    "type":             "myType",
    "fields": {
        "owner":        "root@example.com",
        "priority":     "2"
    }
}`

const snapJSON = `{
    "name":             "Snapshot %03[1]d-%03[2]d",
    "id":               "vfs-%03[1]d-%03[2]d",
    "volumeID":         "vfs-%03[1]d",
    "volumeSize":       10240,
    "status":           "online",
    "startTime":        %[3]d,
    "fields": {
        "owner":        "root@example.com",
        "priority":     "2"
    }
}`
