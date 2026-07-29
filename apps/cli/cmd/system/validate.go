package system

import (
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

 func ValidateSystem() error {
	ui.Step(1, 5, "Checking operating system")

	checkOs()
	// if err := checkOs(); err != nil {
	// 	return err
	// }
	// ui.Success("Operating system supported")

	ui.Step(2, 5, "Checking operating system Archtecture")
	if err := CheckArchitecture(); err != nil {
		return err
	}

	ui.Success("Opearting system archtecture supported")

	ui.Step(3,5,"Checking if Docker engine is running ")
    
	if err := CheckDocker(); err != nil {
		return err
	}

	ui.Success("Docker available")


	ui.Step(4, 5, "Checking Docker Compose")

	if err := CheckDockerCompose(); err != nil {
		return err
	}

	ui.Success("Docker Compose available")


	ui.Step(5, 5, "Checking permissions")

	if err := CheckDockerPermissions(); err != nil {
		return err
	}

	ui.Success("Permissions OK")


	return nil
}

