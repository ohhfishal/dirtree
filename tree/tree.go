package tree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Print(path string, prefix string, depth int, maxDepth int) error {
	if depth > maxDepth {
		panic("High depth not supported yet")
	}

	file, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("getting stat for: %s: %w", path, err)
	}

	if file.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("symlink is not supported: %s", file.Name())
	}

	if file.IsDir() {
		if path != "." {
			fmt.Println(prefix + file.Name() + "/")
		} else {
			fmt.Println(prefix + ".")
		}
		if depth == 0 {
			prefix = "├── "
		} else {
			prefix = "    " + prefix
			prefix = "│" + prefix[1:]
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("getting files: %w", err)
		}

		for i, entry := range entries {
			if i == len(entries)-1 {
				prefix = strings.ReplaceAll(prefix, "├", "└")
			} else {
				prefix = strings.ReplaceAll(prefix, "└", "├")
			}

			Print(
				filepath.Join(path, entry.Name()),
				prefix,
				depth+1,
				maxDepth,
			)
		}
	} else {
		fmt.Println(prefix + file.Name())
	}

	return nil
}
