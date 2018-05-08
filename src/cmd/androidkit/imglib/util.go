package imglib

import (
	"os"
	"path/filepath"
)

/*
 * AndroidXDir is the location of the cache directory, defaults to ~/.AndroidX
 */
var AndroidXDir string

/*
 * Create ~/.AndroidX
 */
func DefaultAndroidXConfigDir() string {
	androidXDefaultDir := ".AndroidX"
	home := os.Getenv("HOME")
	return filepath.Join(home, androidXDefaultDir)
}
