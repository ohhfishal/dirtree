package tree

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CMD struct {
	Path  string `arg:"" optional:"" type:"path" help:"Path to tree from."`
	Depth int    `default:"2" help:"Max depth to recurse."`
}

func (cmd *CMD) Run() error {
	if cmd.Path == "" {
		cmd.Path = "."
	}

	tree, err := findGitFiles(cmd.Path, cmd.Depth)
	if err != nil {
		return fmt.Errorf("finding files: %w", err)
	}
	return tree.Print(os.Stdout)
}

type TreeNode struct {
	Name     string
	Depth    int
	Children map[string]*TreeNode
	Metadata fs.FileInfo
}

func (tree TreeNode) Print(stdout io.Writer) error {
	return tree.print(stdout, "", 0)
}

func (tree TreeNode) print(stdout io.Writer, prefix string, depth int) error {
	// TODO: Buffer this
	fmt.Fprint(stdout, prefix)
	if tree.Metadata.IsDir() {
		if tree.Name != "." {
			fmt.Fprint(stdout, tree.Name)
			fmt.Fprintln(stdout, "/")
		} else {
			fmt.Fprintln(stdout, ".")
		}

		if depth == 0 {
			prefix = "├── "
		} else {
			// This introduces a bug somehow
			// We only want to add the bar if we know there is a child under us
			prefix = "    " + prefix
			prefix = "│" + prefix[1:]
		}

		i := 0
		for _, child := range tree.Children {
			i++
			if i == len(tree.Children) {
				prefix = strings.ReplaceAll(prefix, "├", "└")
			} else {
				prefix = strings.ReplaceAll(prefix, "└", "├")
			}
			child.print(stdout, prefix, depth+1)

		}
	} else {
		fmt.Fprintln(stdout, tree.Metadata.Name())
	}
	return nil
}

func (tree *TreeNode) AddChild(start string, path string) error {
	if tree.Depth <= 0 {
		return fmt.Errorf("can not add child to: %s", tree.Name)
	}

	metadata, err := os.Lstat(filepath.Join(start, path))
	if err != nil {
		return fmt.Errorf("getting status for %s: %w", path, err)
	}

	paths := strings.Split(path, string(filepath.Separator))
	return tree.addChild(start, paths, metadata)
}

func (tree *TreeNode) addChild(parent string, paths []string, metadata fs.FileInfo) error {
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

func findGitFiles(path string, maxDepth int) (*TreeNode, error) {
	metadata, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("getting status root %s: %w", path, err)
	}

	cmd := exec.Command("git", "-C", path, "ls-files", "-o", "-c", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running git command: %w", err)
	}
	lines := strings.Split(string(output), "\n")

	tree := TreeNode{
		Name:     metadata.Name(),
		Metadata: metadata,
		Depth:    maxDepth,
		Children: map[string]*TreeNode{},
	}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if err := tree.AddChild(path, line); err != nil {
			return nil, fmt.Errorf("navigating tree: %w", err)
		}
	}
	return &tree, nil
}

func PrintAll(path string, prefix string, depth int, maxDepth int) error {
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

			PrintAll(
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
