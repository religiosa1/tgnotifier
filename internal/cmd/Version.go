package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"text/tabwriter"

	"github.com/religiosa1/tgnotifier/internal/config"
)

// version can be set at build time using:
// go build -ldflags="-X 'github.com/religiosa1/tgnotifier/internal/cmd.version=v1.2.3'"
// or using Taskfile: task build VERSION='v1.2.3'
var version = ""

type Version struct{}

func (cmd *Version) Run() error {
	fmt.Println(cmd.GetVersion())
	// Show config paths

	fmt.Println()
	fmt.Println("Config paths:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, " User:\t%s\n", config.UserConfigPath)
	fmt.Fprintf(w, " Global:\t%s\n", config.GlobalConfigPath)
	w.Flush()
	return nil
}

func (v Version) GetVersion() string {
	// using version from ldflags first if defined
	if version != "" {
		return version
	}
	// if ldflags versions is not defined, using version from buildinfo
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(no version info available)"
	}

	buildVersion := info.Main.Version

	switch buildVersion {
	case "":
		buildVersion = "(devel) unknown"
	case "(devel)":
		var commit string
		var dirty bool
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commit = setting.Value
			}
			if setting.Key == "vcs.modified" {
				dirty = setting.Value == "true"
			}
		}
		buildVersion += " " + commit
		if dirty {
			buildVersion += " dirty"
		}
	default:
	}

	return buildVersion
}
