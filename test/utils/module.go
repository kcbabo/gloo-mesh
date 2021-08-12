package utils

import (
	"os/exec"
	"strings"
)

// get the file path to a go module
func GetModulePath(modName string) string {
	res, err := exec.Command("go", "list", "-f", "{{ .Dir }}", "-m", modName).CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(res))
}
