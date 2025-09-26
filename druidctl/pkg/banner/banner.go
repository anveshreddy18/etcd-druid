package banner

import (
	"strings"

	"github.com/gardener/etcd-druid/druidctl/pkg/output"
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

	outputService := output.NewService(output.OutputTypeCharm)

	for _, line := range lines {
		outputService.RawHeader(line)
	}

	outputService.RawHeader("Version: " + Version)
}
