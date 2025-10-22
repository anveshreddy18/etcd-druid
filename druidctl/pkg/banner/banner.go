package banner

import (
	"strings"

	"github.com/gardener/etcd-druid/druidctl/pkg/log"
)

var asciiArt = `
▶  ██████╗ ██████╗ ██╗   ██╗██╗██████╗  ██████╗████████╗██╗     
▶  ██╔══██╗██╔══██╗██║   ██║██║██╔══██╗██╔════╝╚══██╔══╝██║     
▶  ██║  ██║██████╔╝██║   ██║██║██║  ██║██║        ██║   ██║     
▶  ██║  ██║██╔══██╗██║   ██║██║██║  ██║██║        ██║   ██║     
▶  ██████╔╝██║  ██║╚██████╔╝██║██████╔╝╚██████╗   ██║   ███████╗
▶  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝╚═════╝  ╚═════╝   ╚═╝   ╚══════╝
`

var Version = "v0.0.1"

func ShowBanner(disableBanner bool) {
	if disableBanner {
		return
	}

	lines := strings.Split(strings.TrimSpace(asciiArt), "\n")

	logger := log.NewLogger(log.LogTypeCharm)

	for _, line := range lines {
		logger.RawHeader(line)
	}

	logger.RawHeader("Version: " + Version)
}
