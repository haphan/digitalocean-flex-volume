package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/digitalocean/cloud"
	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/flex"
)

// VolumePlugin is a Digital Ocean flex volume plugin
type VolumePlugin struct {
	manager *cloud.DigitalOceanManager
}

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
		Message: "DigitalOcean flex driver initialized",
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
		return nil, fmt.Errorf("DigitalOcean volume needs VolumeID property at flex options")
	}

	r := &flex.DriverStatus{
		Status:     flex.StatusSuccess,
		VolumeName: opt.VolumeID,
	}
	return r, nil
}

// Mount volume at the dir where pods will use it
func (v *VolumePlugin) Mount(mountdir string, options string) (*flex.DriverStatus, error) {
	r := &flex.DriverStatus{
		Status:  flex.StatusNotSupported,
		Message: "mount",
	}
	return r, nil
}

// Unmount the volume at mount directory
func (v *VolumePlugin) Unmount(mountdir string) (*flex.DriverStatus, error) {
	r := &flex.DriverStatus{
		Status:  flex.StatusNotSupported,
		Message: "unmount",
	}
	return r, nil
}
