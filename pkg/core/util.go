package core

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofrontier-com/go-utils/output"
)

func removeFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func changedDirs(filesChanged []string, modulesDir string, modulesPath string) []string {
	dirschanged := make([]string, 0)
	for _, fc := range filesChanged {
		output.PrintlnInfo(fc, modulesDir, strings.HasPrefix(fc, modulesDir))

		fcabs, _ := filepath.Abs(fc)
		modulesDirAbs, _ := filepath.Abs(modulesDir)
		output.PrintlnInfo(fcabs, modulesDirAbs)
		if strings.HasPrefix(fcabs, modulesDirAbs) {
			a := strings.Split(fc, "/")
			if len(a) > 2 { // make sure the changed file is of the form [azure resource-group main.tf]
				dirExists := true
				_, err := os.Stat(path.Join(modulesPath, a[1]))
				if os.IsNotExist(err) {
					dirExists = false
				}
				inDirschanged := false
				for _, dir := range dirschanged {
					if dir == a[1] {
						inDirschanged = true
					}
				}
				if inDirschanged == false && dirExists == true {
					dirschanged = append(dirschanged, a[1])
				}
			}
		}
	}

	return dirschanged
}

func getVersion(dir string) (string, error) {
	file, err := os.Open(path.Join(dir, "VERSION"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	retval := strings.TrimSuffix(string(bytes), "\n")

	return retval, nil
}
