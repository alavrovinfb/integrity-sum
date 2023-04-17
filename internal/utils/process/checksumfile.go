package process

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

// Creates checksum file name according to template <image name:tag>.<algorithm>
// e.g nginx:lates.md5
func CheckSumFile(procName, alg string) (string, error) {
	procImage := viper.GetStringMapString("process-image")
	image := procImage[procName]
	ns := viper.GetString("pod-namespace")
	parts := strings.Split(image, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("%s", "incorrect image name")
	}
	return fmt.Sprintf("%s/%s/%s.%s", ns, parts[0], parts[1], strings.ToLower(alg)), nil
}
