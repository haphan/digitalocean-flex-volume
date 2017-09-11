/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	doctx "github.com/digitalocean/godo/context"
	"golang.org/x/oauth2"
)

const (
	godoActionErrored   = "errored"
	godoActionCheckTick = 1000

	// DevicePrefix for Digital Ocean mounts
	DevicePrefix             = "/dev/disk/by-id/scsi-0DO_Volume_"
	dropletRegionMetadataURL = "http://169.254.169.254/metadata/v1/region"
)

// DigitalOceanManager communicates with the DO API
type DigitalOceanManager struct {
	client *godo.Client
	region string
}

// TokenSource represents and oauth2 token source
type tokenSource struct {
	AccessToken string
}

// Token returns an oauth2 token
func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// NewDigitalOceanManager returns a Digitial Ocean manager
func NewDigitalOceanManager(token string) (*DigitalOceanManager, error) {

	if token == "" {
		return nil, errors.New("DigitalOcean token is empty")
	}

	tokenSource := &tokenSource{AccessToken: token}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)

	m := &DigitalOceanManager{
		client: client,
	}

	// generate client and test retrieving account info
	_, err := m.GetAccount()
	if err != nil {
		return nil, err
	}

	return m, nil
}

// GetAccount returns the token related account
func (m *DigitalOceanManager) GetAccount() (*godo.Account, error) {
	account, _, err := m.client.Account.Get(doctx.TODO())
	if err != nil {
		return nil, err
	}
	return account, nil
}

// GetDroplet retrieves the droplet by ID
func (m *DigitalOceanManager) GetDroplet(dropletID int) (*godo.Droplet, error) {
	droplet, _, err := m.client.Droplets.Get(doctx.TODO(), dropletID)
	if err != nil {
		return nil, err
	}
	return droplet, err
}

// DropletList return all droplets
func (m *DigitalOceanManager) DropletList() ([]godo.Droplet, error) {
	list := []godo.Droplet{}
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := m.client.Droplets.List(doctx.TODO(), opt)
		if err != nil {
			return nil, err
		}

		for _, d := range droplets {
			list = append(list, d)
		}
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}
		opt.Page = page + 1
	}
	return list, nil
}

// GetVolume given an unique Digital Ocean identifier returns the volume
func (m *DigitalOceanManager) GetVolume(volumeID string) (*godo.Volume, error) {
	vol, _, err := m.client.Storage.GetVolume(doctx.TODO(), volumeID)
	if err != nil {
		return nil, err
	}
	return vol, nil
}

// GetVolumeByName retrieves a volume given the name
// region will be obtained using this droplet's metadata
func (m *DigitalOceanManager) GetVolumeByName(name string) (*godo.Volume, error) {
	region, err := currentRegion()
	if err != nil {
		return nil, err
	}
	p := &godo.ListVolumeParams{
		Name:   name,
		Region: region,
	}
	vol, _, err := m.client.Storage.ListVolumes(doctx.TODO(), p)
	if err != nil {
		return nil, err
	}

	if len(vol) != 1 {
		return nil, fmt.Errorf("found more than one volume named %q at region %q", name, region)
	}

	return &vol[0], nil
}

// AttachVolumeAndWait attaches volume to given droplet
// it will wait until the attach action is completed
func (m *DigitalOceanManager) AttachVolumeAndWait(volumeID string, dropletID int, timeout time.Duration) error {
	action, _, err := m.client.StorageActions.Attach(doctx.TODO(), volumeID, dropletID)
	if err != nil {
		return err
	}

	err = m.waitVolumeActionCompleted(volumeID, action.ID, timeout)
	if err != nil {
		return err
	}

	return nil
	// DevicePrefix + vol.Name, nil
}

// VolumeNameFromDevicePath given a device path returns a volume name
func (m *DigitalOceanManager) VolumeNameFromDevicePath(device string) (string, error) {
	if !strings.HasPrefix(device, DevicePrefix) {
		return "", fmt.Errorf("device path %q does not seems to be a DigitalOcean volume", device)
	}
	return device[len(DevicePrefix):], nil
}

// DeviceFromVolumeID given a volumeID returns it's device path
func (m *DigitalOceanManager) DeviceFromVolumeID(volumeID string) (string, error) {
	vol, err := m.GetVolume(volumeID)
	if err != nil {
		return "", err
	}
	return DevicePrefix + vol.Name, nil
}

// DetachVolumeAndWait detaches a disk to given droplet
func (m *DigitalOceanManager) DetachVolumeAndWait(volumeID string, dropletID int, timeout time.Duration) error {
	vol, err := m.GetVolume(volumeID)
	if err != nil {
		return err
	}

	needDetach := false
	for _, id := range vol.DropletIDs {
		if id == dropletID {
			needDetach = true
		}
	}

	if needDetach {
		action, _, err := m.client.StorageActions.DetachByDropletID(doctx.TODO(), volumeID, dropletID)
		if err != nil {
			return err
		}

		err = m.waitVolumeActionCompleted(volumeID, action.ID, timeout)
		if err != nil {
			return err
		}
	}

	return nil

}

// FindDropletFromNodeName retrieves the droplet given the kubernetes node name
// Droplet name and Node name should match.
// If not, we will try to match the name with private and public IP
func (m *DigitalOceanManager) FindDropletFromNodeName(node string) (*godo.Droplet, error) {

	// try to find droplet with same name as the kubernetes node
	droplets, err := m.DropletList()
	if err != nil {
		return nil, err
	}

	for _, droplet := range droplets {
		if droplet.Name == node {
			return &droplet, nil
		}
	}

	// Alternative: if not found,
	// Internal or External IP seems to be our safest bet when names doesn't match
	for _, droplet := range droplets {
		ip, err := droplet.PrivateIPv4()
		if err != nil {
			return nil, err
		}
		if ip == node {
			return &droplet, nil
		}
		ip, err = droplet.PublicIPv4()
		if err != nil {
			return nil, err
		}
		if ip == node {
			return &droplet, nil
		}
	}

	return nil, fmt.Errorf("could not match node name to droplet name, private IP or public IP")
}

func (m *DigitalOceanManager) waitVolumeActionCompleted(volumeID string, actionID int, timeout time.Duration) error {
	var lastError error

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()
	ticker := time.NewTicker(time.Second * godoActionCheckTick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			action, _, err := m.client.StorageActions.Get(ctx, volumeID, actionID)
			if err != nil {
				lastError = err
			}

			if action.Status == godo.ActionCompleted {
				return nil
			}
			if action.Status == godoActionErrored {
				return fmt.Errorf("there was and storage action error at DigitalOcean: %s", action.String())
			}
			if action.Status != godo.ActionInProgress {
				return fmt.Errorf("received unexpected action status %q from DigitalOcean", action.Status)
			}
		case <-ctx.Done():
			msg := fmt.Sprintf("storage creation at DigitalOcean for volume %q timed out", volumeID)
			if lastError != nil {
				msg += ": " + lastError.Error()
			}
			return fmt.Errorf(msg)
		}
	}

	// for {
	// 	action, _, err := m.client.StorageActions.Get(doctx.TODO(), volumeID, actionID)
	// 	if err != nil {
	// 		lastError = err
	// 	}
	// 	switch action.Status {
	// 	case godo.ActionCompleted:
	// 		return nil
	// 	case godoActionErrored:
	// 		return fmt.Errorf("There was and storage action error at Digital Ocean: %s", action.String())
	// 	case godo.ActionInProgress:
	// 	default:
	// 		return fmt.Errorf("Received unexpected action status %q from Digital Ocean", action.Status)
	// 	}
	// 	if time.Second*timeout < time.Since(start) {
	// 		errMsg := "Timeout attaching volume " + volumeID
	// 		if lastError != nil {
	// 			errMsg += lastError.Error()
	// 		}
	// 		return fmt.Errorf(errMsg)
	// 	}
	// 	time.Sleep(godoActionTimeSpan * time.Millisecond)
	// }
}

// currentRegion returns the current region for the droplet
func currentRegion() (string, error) {
	resp, err := http.Get(dropletRegionMetadataURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error retrieving droplet region: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
