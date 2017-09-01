package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/digitalocean/cloud"
	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/flex"
)

const (
	digitalOceanAttachTimeout = 200
)

// VolumePlugin is a Digital Ocean flex volume plugin
type VolumePlugin struct {
	manager *cloud.DigitalOceanManager
}

// func debugFile(msg string) {
// 	file := "/var/log/digitalocean.log"
// 	var f *os.File
//
// 	if _, err := os.Stat(file); os.IsNotExist(err) {
// 		f, err = os.Create(file)
// 		if err != nil {
// 			panic(err)
// 		}
// 	} else {
// 		f, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	defer f.Close()
//
// 	if _, err := f.WriteString(msg + "\n"); err != nil {
// 		panic(err)
// 	}
//
// 	// msgb := []byte(msg + "\n")
// 	// ioutil.WriteFile("/tmp/digitalocean.log", msgb, 0644)
// }

// digitalOceanOptions from the flex plugin
type digitalOceanOptions struct {
	// ApiKey string `json:"kubernetes.io/secret/apiKey"`
	FsType         string `json:"kubernetes.io/fsType"`
	PVorVolumeName string `json:"kubernetes.io/pvOrVolumeName"`
	RW             string `json:"kubernetes.io/readwrite"`
	VolumeName     string `json:"volumeName,omitempty"`
	VolumeID       string `json:"volumeID,omitempty"`
}

// NewDigitalOceanVolumePlugin creates a Digital Ocean flex plugin
func NewDigitalOceanVolumePlugin(m *cloud.DigitalOceanManager) flex.VolumePlugin {
	return &VolumePlugin{
		manager: m,
	}
}

// Init driver
func (v *VolumePlugin) Init() (*flex.DriverStatus, error) {
	return &flex.DriverStatus{
		Status:  flex.StatusSuccess,
		Message: "Digital Ocean flex driver initialized",
		Capabilities: &flex.DriverCapabilities{
			Attach:         true,
			SELinuxRelabel: true,
		},
	}, nil
}

func (v *VolumePlugin) newOptions(options string) (*digitalOceanOptions, error) {
	opts := &digitalOceanOptions{}
	if err := json.Unmarshal([]byte(options), opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// GetVolumeName Retrieves a unique volume name
func (v *VolumePlugin) GetVolumeName(options string) (*flex.DriverStatus, error) {
	opt, err := v.newOptions(options)
	if err != nil {
		return nil, err
	}

	if opt.VolumeID == "" {
		return nil, fmt.Errorf("Digital Ocean volume needs VolumeID property")
	}

	r := &flex.DriverStatus{
		Status:     flex.StatusSuccess,
		VolumeName: opt.VolumeID,
	}
	return r, nil
}

// Attach volume to the node
func (v *VolumePlugin) Attach(options string, node string) (*flex.DriverStatus, error) {
	opt, err := v.newOptions(options)
	if err != nil {
		return nil, err
	}

	droplet, err := v.manager.FindDropletFromNodeName(node)
	if err != nil {
		return nil, err
	}

	device, err := v.manager.AttachVolume(opt.VolumeID, droplet.ID, digitalOceanAttachTimeout)
	if err != nil {
		return nil, err
	}

	return &flex.DriverStatus{
		Status:     flex.StatusSuccess,
		DevicePath: device,
	}, nil
}

// Detach the volume from the node
func (v *VolumePlugin) Detach(device, node string) (*flex.DriverStatus, error) {

	volumeID, err := v.manager.VolumeIDFromDevice(device)
	if err != nil {
		return nil, err
	}

	droplet, err := v.manager.FindDropletFromNodeName(node)
	if err != nil {
		return nil, err
	}

	err = v.manager.DetachVolume(volumeID, droplet.ID, digitalOceanAttachTimeout)
	if err != nil {
		return nil, err
	}

	return &flex.DriverStatus{
		Status: flex.StatusSuccess,
	}, nil
}

// WaitForAttach until the volume is attached to the node
// No need to implement since we wait at the Attach command
func (v *VolumePlugin) WaitForAttach(device string, options string) (*flex.DriverStatus, error) {
	r := &flex.DriverStatus{
		Status: flex.StatusNotSupported,
	}
	return r, nil
}

// IsAttached checks for the volume to be attached to the node
func (v *VolumePlugin) IsAttached(options string, node string) (*flex.DriverStatus, error) {
	opt, err := v.newOptions(options)
	if err != nil {
		return nil, err
	}

	found, err := v.manager.FindDropletFromNodeName(node)
	if err != nil {
		return nil, err
	}

	droplet, err := v.manager.GetDroplet(found.ID)
	if err != nil {
		return nil, err
	}

	for _, attachedID := range droplet.VolumeIDs {
		if attachedID == opt.VolumeID {
			return &flex.DriverStatus{
				Status:   flex.StatusSuccess,
				Attached: true,
			}, nil
		}
	}

	return &flex.DriverStatus{
		Status:   flex.StatusSuccess,
		Attached: false,
	}, nil
}

// MountDevice mounts the volume as a device
func (v *VolumePlugin) MountDevice(mountdir, device string, options string) (*flex.DriverStatus, error) {
	r := &flex.DriverStatus{
		Status: flex.StatusNotSupported,
	}
	return r, nil
}

// UnmountDevice from the node
func (v *VolumePlugin) UnmountDevice(device string) (*flex.DriverStatus, error) {
	r := &flex.DriverStatus{
		Status: flex.StatusNotSupported,
	}
	return r, nil
}

// Mount volume at the dir where pods will use it
func (v *VolumePlugin) Mount(mountdir string, options string) (*flex.DriverStatus, error) {
	// opt, err := v.newOptions(options)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// fsType := opt.FsType
	// if fsType == "" {
	// 	return nil, error.New("No filesystem type specified")
	// }

	// args := []string{"-a", source}
	// cmd := mounter.Runner.Command("fsck", args...)
	// out, err := cmd.CombinedOutput()
	// if err != nil {
	// 	ee, isExitError := err.(utilexec.ExitError)
	// 	switch {
	// 	case err == utilexec.ErrExecutableNotFound:
	// 		glog.Warningf("'fsck' not found on system; continuing mount without running 'fsck'.")
	// 	case isExitError && ee.ExitStatus() == fsckErrorsCorrected:
	// 		glog.Infof("Device %s has errors which were corrected by fsck.", source)
	// 	case isExitError && ee.ExitStatus() == fsckErrorsUncorrected:
	// 		return fmt.Errorf("'fsck' found errors on device %s but could not correct them: %s.", source, string(out))
	// 	case isExitError && ee.ExitStatus() > fsckErrorsUncorrected:
	// 		glog.Infof("`fsck` error %s", string(out))
	// 	}
	// }

	// var res unix.Stat_t
	// if err := unix.Stat(device, &res); err != nil {
	// 	return Fail("Could not stat ", device, ": ", err.Error())
	// }
	//
	// if res.Mode&unix.S_IFMT != unix.S_IFBLK {
	// 	return Fail("Not a block device: ", device)
	// }
	//
	// if isMounted(targetDir) {
	// 	return Succeed()
	// }
	//
	// mkfsCmd := exec.Command("mkfs", "-t", fsType, device)
	// if mkfsOut, err := mkfsCmd.CombinedOutput(); err != nil {
	// 	return Fail("Could not mkfs: ", err.Error(), " Output: ", string(mkfsOut))
	// }

	// if err := os.MkdirAll(targetDir, 0750); err != nil {
	// 	return Fail("Could not create directory: ", err.Error())
	// }
	//
	// mountCmd := exec.Command("mount", device, targetDir)
	// if mountOut, err := mountCmd.CombinedOutput(); err != nil {
	// 	return Fail("Could not mount: ", err.Error(), " Output: ", string(mountOut))
	// }
	//
	// return Succeed()

	r := &flex.DriverStatus{
		Status:  flex.StatusNotSupported,
		Message: "mount",
	}
	return r, nil
}

// Unmount the volume at mount directory
func (v *VolumePlugin) Unmount(mountdir string) (*flex.DriverStatus, error) {
	// debugFile("UnMnt")
	// debugFile(fmt.Sprintf("Unmount: mntdir -> %s ", mountdir))
	r := &flex.DriverStatus{
		Status:  flex.StatusSuccess,
		Message: "unmount",
	}
	return r, nil
}
