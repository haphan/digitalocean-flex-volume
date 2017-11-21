package flex

import (
	"encoding/json"
	"fmt"
	"os"
)

// Status codes
const (
	StatusSuccess      = "Success"
	StatusFailure      = "Failure"
	StatusNotSupported = "Not supported"
)

// Flex commands
const (
	initCmd          = "init"
	getVolumeNameCmd = "getvolumename"
	isAttachedCmd    = "isattached"
	attachCmd        = "attach"
	waitForAttachCmd = "waitforattach"
	mountDeviceCmd   = "mountdevice"
	detachCmd        = "detach"
	waitForDetachCmd = "waitfordetach"
	unmountDeviceCmd = "unmountdevice"
	mountCmd         = "mount"
	unmountCmd       = "unmount"
)

// VolumePlugin defines the interface that the internal plugin must implement
type VolumePlugin interface {
	Init() (*DriverStatus, error)
	GetVolumeName(options string) (*DriverStatus, error)
	Attach(options string, node string) (*DriverStatus, error)
	Detach(device, node string) (*DriverStatus, error)
	WaitForAttach(device string, options string) (*DriverStatus, error)
	IsAttached(options string, node string) (*DriverStatus, error)
	MountDevice(mountdir, device string, options string) (*DriverStatus, error)
	UnmountDevice(device string) (*DriverStatus, error)
	Mount(mountdir string, options string) (*DriverStatus, error)
	Unmount(mountdir string) (*DriverStatus, error)
}

// DriverStatus represents the return value of the driver callout.
type DriverStatus struct {
	Status       string              `json:"status"`
	Message      string              `json:"message,omitempty"`
	DevicePath   string              `json:"device,omitempty"`
	VolumeName   string              `json:"volumeName,omitempty"`
	Attached     bool                `json:"attached,omitempty"`
	Capabilities *DriverCapabilities `json:",omitempty"`
}

// DriverCapabilities stores Digital Ocean volume capabilities
type DriverCapabilities struct {
	Attach         bool `json:"attach"`
	SELinuxRelabel bool `json:"selinuxRelabel"`
}

// Command contains all parameters needed to run a plugin operation
type Command struct {
	command  string
	nodeName string
	device   string
	mountdir string
	options  string
}

// NewFlexCommand given an argument list returns a Flex Command structure
func NewFlexCommand(args []string) (*Command, error) {
	if len(args) == 1 {
		return nil, fmt.Errorf("no flex command argument found")
	}

	fc := &Command{command: args[1]}
	fa := args[2:]

	switch fc.command {

	case initCmd:
		return fc, nil

	case getVolumeNameCmd:
		fc.options = fa[0]

	case attachCmd:
		fc.options = fa[0]
		fc.nodeName = fa[1]

	case detachCmd:
		fc.device = fa[0]
		fc.nodeName = fa[1]

	case waitForAttachCmd:
		fc.device = fa[0]
		fc.options = fa[1]

	case isAttachedCmd:
		fc.options = fa[0]
		fc.nodeName = fa[1]

	case mountDeviceCmd:
		fc.mountdir = fa[0]
		fc.device = fa[1]
		fc.options = fa[2]

	case unmountDeviceCmd:
		fc.device = fa[0]

	case mountCmd:
		fc.mountdir = fa[0]
		fc.options = fa[1]

	case unmountCmd:
		fc.mountdir = fa[0]

	default:
		return nil, fmt.Errorf("command %q not recognized as a valid flex command", fc.command)
	}

	return fc, nil
}

// Manager is able to execute flex commands
type Manager struct {
	output *os.File
	plugin VolumePlugin
}

// NewManager returns a Flex manager
func NewManager(plugin VolumePlugin, output *os.File) *Manager {
	return &Manager{
		output: output,
		plugin: plugin,
	}
}

// ExecuteCommand given the command and the plugin
func (m *Manager) ExecuteCommand(fc *Command) (*DriverStatus, error) {
	switch fc.command {
	// case initCmd:
	// 	return m.plugin.Init()
	// case getVolumeNameCmd:
	// 	return m.plugin.GetVolumeName(fc.options)
	case initCmd:
		return m.plugin.Init()
	case attachCmd:
		return m.plugin.Attach(fc.options, fc.nodeName)
	case detachCmd:
		return m.plugin.Detach(fc.device, fc.nodeName)
	case waitForAttachCmd:
		return m.plugin.WaitForAttach(fc.device, fc.options)
	case isAttachedCmd:
		return m.plugin.IsAttached(fc.options, fc.nodeName)
	case mountDeviceCmd:
		return m.plugin.MountDevice(fc.mountdir, fc.device, fc.options)
	case unmountDeviceCmd:
		return m.plugin.UnmountDevice(fc.device)
	case mountCmd:
		return m.plugin.Mount(fc.mountdir, fc.options)
	case unmountCmd:
		return m.plugin.Unmount(fc.mountdir)
	}
	return &DriverStatus{
		Status: StatusNotSupported,
	}, nil
}

// WriteError creates a Flex response containing an error
func (m *Manager) WriteError(e error) {
	ds := &DriverStatus{
		Status:  StatusFailure,
		Message: e.Error(),
	}
	j, err := json.Marshal(ds)
	if err != nil {
		fmt.Printf("could not return JSON encoded error message: %s", err.Error())
		return
	}
	fmt.Fprintln(m.output, string(j))
}

// WriteDriverStatus writes the driver status structure to the output stream
func (m *Manager) WriteDriverStatus(ds *DriverStatus) error {
	j, err := json.Marshal(ds)
	if err != nil {
		return fmt.Errorf("error encoding driver status to JSON: %s", err.Error())
	}

	fmt.Println(string(j))
	return nil
}
