package flex

import (
	"reflect"
	"testing"
)

func TestNexFlexCommand(t *testing.T) {
	cases := []struct {
		args            []string
		expectedCommand *Command
		expectedError   bool
	}{
		{
			[]string{"cmd"},
			nil,
			true,
		},
		{
			[]string{"cmd", "unknownCommand"},
			nil,
			true,
		},
		{
			[]string{"cmd", "init"},
			&Command{
				command: "init",
			},
			false,
		},
		{
			[]string{"cmd", "getvolumename", `{"kubernetes.io/fsType":"ext4","kubernetes.io/pvOrVolumeName":"prueba","kubernetes.io/readwrite":"rw","volumeID":"id0123456789","volumeName":"prueba"}`},
			&Command{
				command: "getvolumename",
				options: `{"kubernetes.io/fsType":"ext4","kubernetes.io/pvOrVolumeName":"prueba","kubernetes.io/readwrite":"rw","volumeID":"id0123456789","volumeName":"prueba"}`,
			},
			false,
		},
	}

	for _, c := range cases {
		cmd, e := NewFlexCommand(c.args)
		if c.expectedError {
			if e == nil {
				t.Errorf("expected error building flex command for arguments %q", c.args)
				continue
			}
		} else {
			if e != nil {
				t.Errorf("an error ocurred building flex command for arguments %q: %s", c.args, e)
				continue
			}
		}

		if cmd == nil && c.expectedCommand == nil {
			continue
		}

		if !reflect.DeepEqual(cmd, c.expectedCommand) {
			t.Errorf("arguments %q expected flex command %+v but got %+v", c.args, c.expectedCommand, cmd)
		}
	}
}
