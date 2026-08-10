//go:build windows

package envpath

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// environmentKey is where the machine-wide environment lives. The per-user
// key under HKCU would need no elevation, but the platform is a machine-wide
// installation and an administrator installing it for the machine should not
// produce a CLI that only their own account can run.
const environmentKey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`

// Broadcasting the change. Without this, the new PATH is in the registry but
// no running process knows: Explorer caches the environment and hands its
// copy to every shell it launches, so "docksight" stays unresolvable in new
// terminals until the next sign-out. WM_SETTINGCHANGE with "Environment" is
// what tells Explorer to re-read it.
const (
	hwndBroadcast   = uintptr(0xFFFF)
	wmSettingChange = uintptr(0x001A)
	smtoAbortIfHung = uintptr(0x0002)
	broadcastWaitMs = uintptr(5000)
)

// Ensure adds directory to the machine PATH, and reports whether it had to.
//
// It is idempotent: a second install finds the entry already there, changes
// nothing and reports false, which is what lets the installer say "already on
// PATH" rather than claiming work it did not do.
func Ensure(directory string) (bool, error) {

	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		environmentKey,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)

	if err != nil {
		return false, fmt.Errorf("failed to open the machine environment: %w", err)
	}

	defer key.Close()

	// The raw value is wanted, not the expanded one. PATH is normally
	// REG_EXPAND_SZ and legitimately contains entries like %SystemRoot%;
	// writing back an expanded copy would bake this machine's current values
	// into a variable that is supposed to stay symbolic.
	current, valueType, err := key.GetStringValue("Path")

	if err != nil && err != registry.ErrNotExist {
		return false, fmt.Errorf("failed to read the machine PATH: %w", err)
	}

	updated, changed := Append(current, directory)

	if !changed {
		return false, nil
	}

	if valueType == registry.EXPAND_SZ || valueType == 0 {
		err = key.SetExpandStringValue("Path", updated)
	} else {
		err = key.SetStringValue("Path", updated)
	}

	if err != nil {
		return false, fmt.Errorf("failed to update the machine PATH: %w", err)
	}

	broadcastEnvironmentChange()

	return true, nil
}

// broadcastEnvironmentChange tells running processes to re-read the
// environment. Failure is not reported: the registry write has already
// succeeded, the PATH is correct from the next sign-in regardless, and a
// hung top-level window is not a reason to fail an install that worked.
func broadcastEnvironmentChange() {

	message, err := windows.UTF16PtrFromString("Environment")

	if err != nil {
		return
	}

	procSendMessageTimeout.Call(
		hwndBroadcast,
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(message)),
		smtoAbortIfHung,
		broadcastWaitMs,
		0,
	)
}

var procSendMessageTimeout = windows.NewLazySystemDLL("user32.dll").
	NewProc("SendMessageTimeoutW")
