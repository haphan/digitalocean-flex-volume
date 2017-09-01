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
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

const (
	godoActionErrored  = "errored"
	godoActionTimeSpan = 1000
	devicePrefix       = "/dev/disk/by-id/scsi-0DO_Volume_"
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
		return nil, errors.New("Digital Ocean token must be informed")
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
	account, _, err := m.client.Account.Get(context.TODO())
	if err != nil {
		return nil, err
	}
	return account, nil
}

// GetDroplet retrieves the droplet by ID
func (m *DigitalOceanManager) GetDroplet(dropletID int) (*godo.Droplet, error) {
	droplet, _, err := m.client.Droplets.Get(context.TODO(), dropletID)
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
		droplets, resp, err := m.client.Droplets.List(context.TODO(), opt)
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

func (m *DigitalOceanManager) GetVolume(volumeID string) (*godo.Volume, error) {
	vol, _, err := m.client.Storage.GetVolume(context.TODO(), volumeID)
	if err != nil {
		return nil, err
	}
	return vol, nil
}

// AttachVolume attaches volume to given droplet
// returns the path the disk is being attached to
func (m *DigitalOceanManager) AttachVolume(volumeID string, dropletID int, timeout int) (string, error) {
	vol, err := m.GetVolume(volumeID)
	if err != nil {
		return "", err
	}

	needAttach := true
	for id := range vol.DropletIDs {
		if id == dropletID {
			needAttach = false
		}
	}

	if needAttach {
		action, _, err := m.client.StorageActions.Attach(context.TODO(), volumeID, dropletID)
		if err != nil {
			return "", err
		}

		err = m.waitVolumeActionCompleted(volumeID, action.ID, timeout)
		if err != nil {
			return "", err
		}
	}
	return devicePrefix + vol.Name, nil
}

// VolumeIDFromDevice given a device path returns a volume ID
func (m *DigitalOceanManager) VolumeIDFromDevice(device string) (string, error) {
	if !strings.HasPrefix(device, devicePrefix) {
		return "", fmt.Errorf("device path %q does not seems to be a Digital Ocean volume", device)
	}
	return device[len(devicePrefix):], nil
}

// DetachVolume detaches a disk to given droplet
func (m *DigitalOceanManager) DetachVolume(volumeID string, dropletID int, timeout int) error {
	vol, err := m.GetVolume(volumeID)
	if err != nil {
		return err
	}

	needDetach := false
	for id := range vol.DropletIDs {
		if id == dropletID {
			needDetach = true
		}
	}

	if needDetach {
		action, _, err := m.client.StorageActions.DetachByDropletID(context.TODO(), volumeID, dropletID)
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

	return nil, fmt.Errorf("Couldn't match node name to droplet name, private IP or public IP")
}

func (m *DigitalOceanManager) waitVolumeActionCompleted(volumeID string, actionID int, timeout int) error {
	var lastError error
	start := time.Now()

	for {
		action, _, err := m.client.StorageActions.Get(context.TODO(), volumeID, actionID)
		if err != nil {
			lastError = err
		}
		switch action.Status {
		case godo.ActionCompleted:
			return nil
		case godoActionErrored:
			return fmt.Errorf("There was and storage action error at Digital Ocean: %s", action.String())
		case godo.ActionInProgress:
		default:
			return fmt.Errorf("Received unexpected action status %q from Digital Ocean", action.Status)
		}
		if time.Second*time.Duration(timeout) > time.Since(start) {
			errMsg := "Timeout attaching volume " + volumeID
			if lastError != nil {
				errMsg += lastError.Error()
			}
			return fmt.Errorf(errMsg)
		}
		time.Sleep(godoActionTimeSpan * time.Millisecond)
	}
}
