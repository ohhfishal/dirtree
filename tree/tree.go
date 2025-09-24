package tree

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type CMD struct {
	Path      string   `arg:"" optional:"" type:"path" help:"Path to use as the tree root."`
	Depth     int      `short:"D" default:"2" help:"Max depth to recurse."`
	FileMode  FileMode `short:"F" enum:"git,file,auto" default:"auto" help:"How to discover files (git, file, auto)." env:"FILE_MODE"`
	ColorMode string   `short:"C" enum:"auto,never,always" default:"auto" help:"When to use colors (auto,always,never)." env:"COLOR_MODE"`
}

func (cmd *CMD) Run(_ context.Context) error {
	if cmd.Path == "" {
		cmd.Path = "."
	}

	if cmd.FileMode == FileModeAutomatic {
		cmd.FileMode = InferFileMode(cmd.Path)
	}

	var tree *TreeNode
	var err error
	switch cmd.FileMode {
	case FileModeGit:
		tree, err = findGitFiles(cmd.Path, cmd.Depth)
		if err != nil {
			return fmt.Errorf("finding files: %w", err)
		}
	case FileModeFile:
		tree, err = findAllFiles(cmd.Path, cmd.Depth)
		if err != nil {
			return fmt.Errorf("finding files: %w", err)
		}
	default:
		return fmt.Errorf("unknown file mode: %s", cmd.FileMode)
	}

	var colors Colors
	switch cmd.ColorMode {
	case "auto":
		colors = ColorsAuto(os.Stdout, os.Getenv)
	case "always":
		colors = ColorsAlways(os.Getenv)
	case "never":
		colors = Colors{}
	default:
		return fmt.Errorf("unknown color mode: %s", cmd.ColorMode)
	}

	// TODO: Print a message like:
	// "%d directories, %d files"
	return tree.Print(os.Stdout, colors)
}

type TreeNode struct {
	Name         string
	Depth        int
	OmitChildren bool
	Children     map[string]*TreeNode
	Metadata     fs.FileInfo
}

func (tree TreeNode) Print(stdout io.Writer, colors Colors) error {
	// TODO: Make colors optional
	return tree.print(stdout, colors, "", 0)
}

func (tree TreeNode) print(stdout io.Writer, colors Colors, prefix string, depth int) error {
	// TODO: Buffer this
	fmt.Fprint(stdout, prefix)
	if tree.Metadata.IsDir() {
		if tree.Name != "." {
			fmt.Fprint(stdout, colors.Format(tree.Name, tree.Metadata))
			fmt.Fprintln(stdout, "/")
		} else {
			fmt.Fprintln(stdout, ".")
		}

		prefix = strings.Replace(prefix, "├──", "│  ", 1)
		prefix = strings.Replace(prefix, "└──", "   ", 1)

		childPrefix := prefix + "├── "
		orphanPrefix := prefix + "└── "

		if tree.OmitChildren {
			fmt.Fprintln(stdout, orphanPrefix+"...")
			return nil
		}

		i := 0
		for _, child := range tree.Children {
			if i == len(tree.Children)-1 {
				child.print(stdout, colors, orphanPrefix, depth+1)
			} else {
				child.print(stdout, colors, childPrefix, depth+1)
			}
			i++
		}
	} else {

		fmt.Fprintln(stdout, colors.Format(tree.Name, tree.Metadata))
	}
	return nil
}

func (tree *TreeNode) AddChild(start string, path string) error {
	metadata, err := os.Lstat(filepath.Join(start, path))
	if err != nil {
		return fmt.Errorf("getting status for %s: %w", path, err)
	}

	paths := strings.Split(path, string(filepath.Separator))
	return tree.addChild(start, paths, metadata)
}

func (tree *TreeNode) addChild(parent string, paths []string, metadata fs.FileInfo) error {
	if tree.Depth <= 0 {
		tree.OmitChildren = true
		return nil
	}

	if len(paths) == 0 {
		// Base case nothing to do
		return nil
	} else if len(paths) == 1 {
		// At a single file or directory
		tree.Children[metadata.Name()] = &TreeNode{
			Name:     metadata.Name(),
			Metadata: metadata,
			Depth:    tree.Depth - 1,
			Children: map[string]*TreeNode{},
		}
	} else if _, ok := tree.Children[paths[0]]; ok {
		// Child already exists
		if err := tree.Children[paths[0]].addChild(filepath.Join(parent, paths[0]), paths[1:], metadata); err != nil {
			return err
		}
	} else {
		// Child does not exist
		nodeMetadata, err := os.Lstat(filepath.Join(parent, paths[0]))
		if err != nil {
			return fmt.Errorf("getting status new node %s: %w", paths[0], err)
		}
		child := TreeNode{
			Name:     paths[0],
			Metadata: nodeMetadata,
			Depth:    tree.Depth - 1,
			Children: map[string]*TreeNode{},
		}

		if err := child.addChild(filepath.Join(parent, paths[0]), paths[1:], metadata); err != nil {
			return err
		}
		tree.Children[paths[0]] = &child
	}
	return nil
}
