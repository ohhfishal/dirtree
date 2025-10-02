package tree

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Stdout interface {
	Stat() (os.FileInfo, error)
}

type Colors struct {
	Directory  string
	Executable string
	Link       string
	Reset      string
	extensions map[string]string
}

func ColorsAuto(stdout Stdout, env func(string) string) Colors {
	if info, err := stdout.Stat(); err == nil {
		mode := info.Mode()
		if mode&fs.ModeNamedPipe != 0 || mode&fs.ModeCharDevice == 0 {
			// Pipe or redirected to file
			return Colors{}
		}
	} else {
		// Disabling since stdout appears broken
		return Colors{}
	}
	return ColorsAlways(env)
}

func ColorsAlways(env func(string) string) Colors {
	if colorsVar := env("LS_COLORS"); colorsVar != "" {
		colors, err := ParseLinuxColors(colorsVar)
		if err != nil {
			slog.Warn("invalid $LS_COLORS", "error", err)
		} else {
			return *colors
		}
	}
	return ColorsDefault()
}

func ColorsDefault() Colors {
	return Colors{
		Directory:  "\033[01;34m",
		Reset:      "\033[0m",
		extensions: map[string]string{}, // Implict enabled flag
	}
}

func (colors Colors) Format(text string, info os.FileInfo) string {
	// Ignore those that are not set properly
	if colors.extensions == nil {
		return text
	}

	// TODO: Does not work with non . based extensions (stdlib ignores them)
	//       At least on my machine is *# and *~ files
	if color, ok := colors.extensions[filepath.Ext(info.Name())]; ok {
		return color + text + colors.Reset
	}

	mode := info.Mode()
	switch {
	// TODO: Do all the other formatting :)
	case info.IsDir():
		return colors.Directory + text + colors.Reset
	case mode.IsRegular() && (mode&0111 != 0):
		return colors.Executable + text + colors.Reset
	case mode&fs.ModeSymlink != 0:
		return colors.Link + text + colors.Reset
	default:
		return text
	}
}

func ParseLinuxColors(lsColors string) (*Colors, error) {
	var colors Colors
	colors.extensions = map[string]string{}
	lines := strings.Split(lsColors, ":")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line: %s: no =", line)
		}

		colorFmt := "\033[%sm"
		// https://man7.org/linux/man-pages/man5/dir_colors.5.html
		// https://github.com/coreutils/coreutils/blob/master/src/dircolors.c
		// TODO: Ensure the rest of the string is valid?
		switch parts[0] {
		case "no":
			// Normal
		case "rs":
			// Reset
			colors.Reset = fmt.Sprintf(colorFmt, parts[1])
		case "di":
			// Directory
			colors.Directory = fmt.Sprintf(colorFmt, parts[1])
		case "ln":
			// Link
			colors.Link = fmt.Sprintf(colorFmt, parts[1])
		case "mh":
			// Multihardlink
		case "pi":
			// Pipe
		case "so":
			// Sock
		case "do":
			// Door
		case "bd":
			// Block
		case "cd":
			// Character or device driver
		case "or":
			// Orphan
		case "mi":
			// Missing
		case "su":
			// Setup ID
		case "sg":
			// Set GID
		case "ca":
			// Regular file w/ capability
		case "tw":
			// Sticky writeable
		case "ow":
			// Other writeable
		case "st":
			// Sticky
		case "ex":
			// Regular file w/ executeable permmission
			colors.Executable = fmt.Sprintf(colorFmt, parts[1])
		case "lc":
			// Left code?
		case "rc":
			// Right code?
		case "ec":
			// End code?
		case "cl":
			// CLRTOEOL ?
		default:
			if !strings.HasPrefix(parts[0], "*") {
				return nil, fmt.Errorf("unknown line: %s", line)
			}
			colors.extensions[parts[0][1:]] = fmt.Sprintf(colorFmt, parts[1])
		}
	}
	return &colors, nil
}
