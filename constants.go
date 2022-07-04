package template

import (
	"os"
	"path/filepath"

	"github.com/mehdi-roozitalab/core_utils"
)

var startDir = func() string {
	wd, _ := os.Getwd()
	wd, _ = core_utils.AbsolutePath(wd)
	return wd
}()
var appLocation, appFolder = func() (string, string) {
	appLocation, _ := os.Executable()
	appLocation, _ = core_utils.AbsolutePath(appLocation)
	return appLocation, filepath.Dir(appLocation)
}()
