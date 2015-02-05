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
	sourceDir     string
	sourcePattern string
	recursive     bool
	destDir       string
	uri           string

	isFileAsset bool

	re       *regexp.Regexp
	destPath string

	sourcePaths []string
	destMTime   time.Time
	sourceMTime time.Time
}

// NewFileAsset creates new asset definition for a single file. The asset is built from the files that match the sourcePattern
// and located in the sourceDir. The result will be written in the destDir directory with the uri path (destDir + uri).
func NewFileAsset(sourceDir string, sourcePattern string, recursive bool, destDir string, uri string) *Asset {
	asset := &Asset{sourceDir: sourceDir, sourcePattern: sourcePattern, recursive: recursive, destDir: destDir, uri: uri}
	asset.isFileAsset = true

	asset.init()
	return asset
}

// NewFolderAsset creates new asset definition for a directory of files.
func NewFolderAsset(sourceDir string, sourcePattern string, recursive bool, destDir string, uri string) *Asset {
	asset := &Asset{sourceDir: sourceDir, sourcePattern: sourcePattern, recursive: recursive, destDir: destDir, uri: uri}
	asset.isFileAsset = false

	asset.init()
	return asset
}

func (p *Asset) init() {
	p.destPath = path.Join(p.destDir, filepath.FromSlash(p.uri))

	re, reErr := regexp.Compile(p.sourcePattern)
	if reErr != nil {
		log.Fatalf("Invalid sourcePattern '%s': %v", p.sourcePattern, reErr)
	}

	p.re = re
}

func (p *Asset) rescan() {
	paths := []string{}

	var mtime time.Time
	filepath.Walk(p.sourceDir, func(path string, info os.FileInfo, pathErr error) error {
		if info == nil {
			log.Fatalf("Invalid asset path '%s'\n%v", path, pathErr)
			return pathErr
		}

		if !p.recursive && info.IsDir() && p.sourceDir != path {
			return filepath.SkipDir
		}

		if !info.IsDir() && p.re.MatchString(path) {
			paths = append(paths, path)

			if info.ModTime().After(mtime) {
				mtime = info.ModTime()
			}
		}

		return nil
	})

	p.sourcePaths = paths

	// Since file asset could be built from many different files in the source directory we are going to
	// include all of the to see if the asset must be rebuilt
	if p.isFileAsset {
		filepath.Walk(p.sourceDir, func(path string, info os.FileInfo, pathErr error) error {
			if info == nil {
				return pathErr
			}

			if info.ModTime().After(mtime) {
				mtime = info.ModTime()
			}

			return nil
		})
	}

	p.sourceMTime = mtime

	fileInfo, err := os.Stat(p.destPath)
	if err == nil && fileInfo.Size() > 0 {
		p.destMTime = fileInfo.ModTime()
	}
}

// MatchesRequestURI function returns true if requestURI matches the uri of the asset.
func (p *Asset) MatchesRequestURI(requestURI string) bool {
	assetURI := strings.TrimPrefix(p.uri, "/")
	rURI := strings.TrimPrefix(requestURI, "/")

	if p.isFileAsset {
		return assetURI == rURI
	}

	return assetURI == "" || strings.HasPrefix(rURI, assetURI)
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
