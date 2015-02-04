// Package pipeline provides simple asset pipeline for Go web apps.
package pipeline

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Asset struct {
	sourceDir string
	pattern   string
	recursive bool
	destDir   string
	uri       string

	re       *regexp.Regexp
	destPath string

	sourcePaths []string
	destMTime   time.Time
	sourceMTime time.Time
}

// NewAsset creates new asset definition. The asset is built from the files that match the sourcePattern
// and located in the sourceDir. The result will be written in the destDir directory with the uri path (destDir + uri).
func NewAsset(sourceDir string, sourcePattern string, recursive bool, destDir string, uri string) *Asset {
	asset := &Asset{sourceDir: sourceDir, pattern: sourcePattern, recursive: recursive, destDir: destDir, uri: uri}
	asset.destPath = path.Join(destDir, filepath.FromSlash(uri))

	re, reErr := regexp.Compile(sourcePattern)
	if reErr != nil {
		log.Fatalf("Invalid sourcePattern '%s': %v", sourcePattern, reErr)
	}

	asset.re = re
	return asset
}

func (p *Asset) rescan() {
	paths := []string{}

	filepath.Walk(p.sourceDir, func(path string, info os.FileInfo, _ error) error {
		if !p.recursive && info.IsDir() && p.sourceDir != path {
			return filepath.SkipDir
		}

		if !info.IsDir() && p.re.MatchString(path) {
			paths = append(paths, path)
		}

		return nil
	})

	p.sourcePaths = paths

	var mtime time.Time
	filepath.Walk(p.sourceDir, func(path string, info os.FileInfo, _ error) error {
		if info.ModTime().After(mtime) {
			mtime = info.ModTime()
		}

		return nil
	})

	p.sourceMTime = mtime

	fileInfo, err := os.Stat(p.destPath)
	if err == nil {
		p.destMTime = fileInfo.ModTime()
	}
}

// MatchesRequestURI function returns true if requestURI matches the uri of the asset.
func (p *Asset) MatchesRequestURI(requestURI string) bool {
	return strings.TrimPrefix(p.uri, "/") == strings.TrimPrefix(requestURI, "/")
}

func (p *Asset) MkDestDir() {
	dirErr := os.MkdirAll(path.Dir(p.destPath), os.ModeDir|os.ModePerm)
	if dirErr != nil {
		log.Fatal("Failed to create directory ", p.destDir, dirErr)
	}
}

// NeedsUpdate function returns true if the asset is older than the files in the source directory and must be regenerated
func (p *Asset) NeedsUpdate() bool {
	p.rescan()
	return p.sourceMTime.After(p.destMTime)
}
