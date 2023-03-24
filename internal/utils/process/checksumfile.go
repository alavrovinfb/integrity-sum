package process

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

// Creates checksum file name according to template <image name:tag>.<algorithm>
// e.g nginx:lates.md5
func CheckSumFile(procName, alg string) string {
	procImage := viper.GetStringMapString("process-image")
	image := procImage[procName]

	return fmt.Sprintf("%s.%s", image, strings.ToLower(alg))
}
