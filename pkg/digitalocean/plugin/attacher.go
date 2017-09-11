package plugin

import (
	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/digitalocean/cloud"
	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/flex"
)

const (
	digitalOceanAttachTimeout = 200
)

// Attach volume to the node
func (v *VolumePlugin) Attach(options string, node string) (*flex.DriverStatus, error) {
	opt, err := v.newOptions(options)
	if err != nil {
		return nil, err
	}

	d, err := v.manager.FindDropletFromNodeName(node)
	if err != nil {
		return nil, err
	}

	// we need to retrieve the droplet to get the volumes (previous call lacks volumes)
	droplet, err := v.manager.GetDroplet(d.ID)
	if err != nil {
		return nil, err
	}

	needAttach := true
	for _, attachedID := range droplet.VolumeIDs {
		if attachedID == opt.VolumeID {
			needAttach = false
		}
	}

	vol, err := v.manager.GetVolume(opt.VolumeID)
	if err != nil {
		return nil, err
	}

	if needAttach {
		err := v.manager.AttachVolumeAndWait(opt.VolumeID, droplet.ID, digitalOceanAttachTimeout)
		if err != nil {
			return nil, err
		}
	}

	return &flex.DriverStatus{
		Status:     flex.StatusSuccess,
		DevicePath: cloud.DevicePrefix + vol.Name,
	}, nil
}

// Detach the volume from the node
func (v *VolumePlugin) Detach(device, node string) (*flex.DriverStatus, error) {

	vol, err := v.manager.GetVolumeByName(device)
	if err != nil {
		return nil, err
	}

	d, err := v.manager.FindDropletFromNodeName(node)
	if err != nil {
		return nil, err
	}

	// we need to retrieve the droplet to get the volumes (previous call lacks volumes)
	droplet, err := v.manager.GetDroplet(d.ID)
	if err != nil {
		return nil, err
	}

	needDetach := false
	for _, attachedID := range droplet.VolumeIDs {
		if attachedID == vol.ID {
			needDetach = true
		}
	}

	if needDetach {
		err = v.manager.DetachVolumeAndWait(vol.ID, droplet.ID, digitalOceanAttachTimeout)
		if err != nil {
			return nil, err
		}
	}

	return &flex.DriverStatus{
		Status: flex.StatusSuccess,
	}, nil
}

// WaitForAttach no need to implement since we wait at the Attach command
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

	d, err := v.manager.FindDropletFromNodeName(node)
	if err != nil {
		return nil, err
	}

	// we need to retrieve the droplet to get the volumes (previous call lacks volumes)
	droplet, err := v.manager.GetDroplet(d.ID)
	if err != nil {
		return nil, err
	}

	isAttached := false
	for _, attachedID := range droplet.VolumeIDs {
		if attachedID == opt.VolumeID {
			isAttached = true
			break
		}
	}

	return &flex.DriverStatus{
		Status:   flex.StatusSuccess,
		Attached: isAttached,
	}, nil
}
