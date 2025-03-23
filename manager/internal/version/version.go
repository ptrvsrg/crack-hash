package version

import (
	"fmt"
	"runtime"
)

var (
	AppVersion = "0.0.0"
	GoVersion  = runtime.Version()
	Platform   = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)
