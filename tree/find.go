package tree

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type FileMode string

const (
	FileModeGit       FileMode = "git"
	FileModeFile      FileMode = "file"
	FileModeAutomatic FileMode = "auto"
)

func InferFileMode(path string) FileMode {
	cmd := exec.Command("git", "-C", path, "status")
	if err := cmd.Run(); err != nil {
		slog.Debug(
			"failed to git status, falling back to File",
			slog.Any("error", err),
		)
		return FileModeFile
	}
	return FileModeGit
}

func findAllFiles(root string, maxDepth int) (*TreeNode, error) {
	metadata, err := os.Lstat(root)
	if err != nil {
		return nil, fmt.Errorf("getting status root %s: %w", root, err)
	}

	tree := TreeNode{
		Name:     metadata.Name(),
		Metadata: metadata,
		Depth:    maxDepth,
		Children: map[string]*TreeNode{},
	}

	// TODO: This appears to be walking it relative
	if err := filepath.WalkDir(root, func(path string, current fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("looking up path: %w", err)
		}

		relative := strings.TrimPrefix(path, root)
		relative = strings.TrimPrefix(relative, string(os.PathSeparator))

		metadata, err := current.Info()
		if err != nil {
			return fmt.Errorf("getting file info: %w", err)
		}

		paths := strings.Split(relative, string(filepath.Separator))

		// +1 to allow the leaf to know if they have children to omit
		if len(paths) > maxDepth+1 {
			return fs.SkipDir
		}

		if err := tree.addChild(root, paths, metadata); err != nil {
			return fmt.Errorf("caching file: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("traversing files: %w", err)
	}
	return &tree, nil
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
