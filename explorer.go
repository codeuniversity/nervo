package nervo

import (
	"io/ioutil"
	"path"
	"strings"
)

var (
	sourceDirectories = []string{"/dev"}
	matchers          = []string{"tty.usb", "ttyACM"}
)

func discoverAttachedControllers() (controllerPorts []string, err error) {
	for _, source := range sourceDirectories {
		files, err := ioutil.ReadDir(source)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			for _, matcher := range matchers {
				if strings.Contains(file.Name(), matcher) {
					controllerPorts = append(controllerPorts, path.Join(source, file.Name()))
				}
			}
		}
	}
	return
}
