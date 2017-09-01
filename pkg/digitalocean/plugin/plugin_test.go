package plugin

import (
	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/flex"

	"reflect"
	"testing"
)

func TestGetVolumeName(t *testing.T) {
	cases := []struct {
		options        string
		expectedStatus *flex.DriverStatus
		expectedError  bool
	}{
		{
			"",
			nil,
			true,
		},
		{
			`{"kubernetes.io/fsType":"ext4","kubernetes.io/pvOrVolumeName":"prueba","kubernetes.io/readwrite":"rw","volumeID":"","volumeName":"prueba"}`,
			nil,
			true,
		},
		{
			`{"kubernetes.io/fsType":"ext4","kubernetes.io/pvOrVolumeName":"prueba","kubernetes.io/readwrite":"rw","volumeID":"id0123456789","volumeName":"prueba"}`,
			&flex.DriverStatus{
				Status:     flex.StatusSuccess,
				VolumeName: "id0123456789",
			},
			false,
		},
	}

	for _, c := range cases {
		vp := &VolumePlugin{}
		ds, e := vp.GetVolumeName(c.options)

		if c.expectedError {
			if e == nil {
				t.Errorf("expected error getting volume name for options %q", c.options)
				continue
			}
		} else {
			if e != nil {
				t.Errorf("an error ocurred getting volume name for options %q: %s", c.options, e)
				continue
			}
		}

		if ds == nil && c.expectedStatus == nil {
			continue
		}

		if !reflect.DeepEqual(ds, c.expectedStatus) {
			t.Errorf("options %q expected volume name %+v but got %+v", c.options, c.expectedStatus, ds)
		}
	}
}
