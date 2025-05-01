package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	// TODO: use param from tool
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Build the directory tree
	root, err := BuildDirectoryTree(cwd)
	if err != nil {
		fmt.Printf("Error building directory tree: %v\n", err)
		panic(err)
	}

	// Get the directory tree as a string
	treeOutput := PrintDirectoryTree(root)

	// Write the output to a file
	err = os.WriteFile("structure.txt", []byte(treeOutput), 0644)
	if err != nil {
		slog.Error("Failed to write structure file", "error", err)
		panic(err)
	}

	slog.Info("Successfully wrote directory structure to structure.txt")

	slog.Info("structure", "s", PrintDirectoryTree(root))
}

// DirectoryEntry represents a file or directory in the structure
type DirectoryEntry struct {
	Name     string
	IsDir    bool
	Children []*DirectoryEntry
}

// BuildDirectoryTree builds a directory tree from the given path
func BuildDirectoryTree(path string) (*DirectoryEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	entry := &DirectoryEntry{
		Name:  filepath.Base(path),
		IsDir: info.IsDir(),
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
		}

		for _, fileInfo := range entries {
			childPath := filepath.Join(path, fileInfo.Name())

			// Skip hidden files/directories (starting with .)
			if strings.HasPrefix(fileInfo.Name(), ".") {
				continue
			}

			var childEntry *DirectoryEntry

			if fileInfo.IsDir() {
				// Recurse into subdirectories
				childEntry, err = BuildDirectoryTree(childPath)
				if err != nil {
					return nil, err
				}
			} else {
				// Add file
				childEntry = &DirectoryEntry{
					Name:  fileInfo.Name(),
					IsDir: false,
				}
			}

			entry.Children = append(entry.Children, childEntry)
		}

		// Sort children (directories first, then files alphabetically)
		sort.Slice(entry.Children, func(i, j int) bool {
			if entry.Children[i].IsDir != entry.Children[j].IsDir {
				return entry.Children[i].IsDir
			}
			return entry.Children[i].Name < entry.Children[j].Name
		})
	}

	return entry, nil
}

// PrintDirectoryTree prints the directory tree to a string with proper formatting
func PrintDirectoryTree(root *DirectoryEntry) string {
	var sb strings.Builder

	// Print the root directory with absolute path first
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}

	sb.WriteString(fmt.Sprintf("Project Structure for: %s\n", cwd))

	// Print the root directory name
	if root.IsDir {
		sb.WriteString(root.Name + "/\n")
	} else {
		sb.WriteString(root.Name + "\n")
	}

	// Print all children with proper indentation
	printSubTree(&sb, root.Children, "")

	return sb.String()
}

// printSubTree recursively prints the subtree with proper indentation and tree lines
func printSubTree(
	sb *strings.Builder,
	entries []*DirectoryEntry,
	prefix string,
) {
	for i, entry := range entries {
		isLastEntry := i == len(entries)-1

		// Print the current entry
		if isLastEntry {
			sb.WriteString(prefix + "└── ")
		} else {
			sb.WriteString(prefix + "├── ")
		}

		// Add directory indicator
		if entry.IsDir {
			sb.WriteString(entry.Name + "/\n")
		} else {
			sb.WriteString(entry.Name + "\n")
		}

		// Recursively print children with updated prefix
		if entry.IsDir && len(entry.Children) > 0 {
			var newPrefix string
			if isLastEntry {
				newPrefix = prefix + "    "
			} else {
				newPrefix = prefix + "│   "
			}
			printSubTree(sb, entry.Children, newPrefix)
		}
	}
}
