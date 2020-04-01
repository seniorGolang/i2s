package utils

import "os/exec"

func GoImports(path string) (err error) {

	var execPath string

	if execPath, err = exec.LookPath("goimports"); err != nil {
		return
	}
	err = exec.Command(execPath, "-w", path).Run()

	return
}
