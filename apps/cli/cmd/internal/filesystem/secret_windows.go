//go:build windows

package filesystem

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// ProtectSecret restricts a file to Administrators and SYSTEM.
//
// Windows ignores the Unix mode bits Go passes to os.OpenFile, so a .env
// created with 0600 is not protected by anything: it inherits the ACL of its
// parent directory. The parent here is under ProgramData, whose default ACL
// grants Users read — which would make POSTGRES_PASSWORD and JWT_SECRET
// readable by every account on the machine, including unprivileged ones and
// anything running as them.
//
// The fix is an explicit DACL with inheritance switched off. PROTECTED is the
// load-bearing flag: without it the two entries below would be added to the
// inherited ones rather than replacing them, and Users would keep its read.
func ProtectSecret(path string) error {

	administrators, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)

	if err != nil {
		return fmt.Errorf("failed to resolve the Administrators group: %w", err)
	}

	localSystem, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)

	if err != nil {
		return fmt.Errorf("failed to resolve the SYSTEM account: %w", err)
	}

	// SYSTEM is granted access alongside Administrators because services and
	// scheduled tasks run as it, and a file only Administrators can read is a
	// file a service cannot.
	dacl, err := windows.ACLFromEntries(
		[]windows.EXPLICIT_ACCESS{
			fullControl(administrators),
			fullControl(localSystem),
		},
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to build an access list for %s: %w", path, err)
	}

	err = windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		dacl,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to restrict access to %s: %w", path, err)
	}

	return nil
}

func fullControl(sid *windows.SID) windows.EXPLICIT_ACCESS {

	return windows.EXPLICIT_ACCESS{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.GRANT_ACCESS,

		// A file has nothing to inherit the entry, and saying so keeps the
		// ACL exactly two entries long and easy to read in Explorer.
		Inheritance: windows.NO_INHERITANCE,

		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_GROUP,
			TrusteeValue: windows.TrusteeValueFromSID(sid),
		},
	}
}
