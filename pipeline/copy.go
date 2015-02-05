package pipeline

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func copyfile(destpath, srcpath string) (written int64, err error) {
	if srcpath == destpath {
		return 0, errors.New("Cannot copy file, source and destination are the same.")
	}

	src, err := os.Open(srcpath)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := os.Create(destpath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	return io.Copy(dst, src)
}

func (p *Asset) performCopyFile(destPath string, sourcePath string, info os.FileInfo) {
	dirErr := os.MkdirAll(filepath.Dir(destPath), os.ModeDir|os.ModePerm)
	if dirErr != nil {
		log.Fatal("Failed to create directory ", destPath, dirErr)
	}

	_, copyErr := copyfile(destPath, sourcePath)
	if copyErr != nil {
		log.Fatalf("Failed to copy '%s' to '%s': %v", sourcePath, destPath, copyErr)
	}

	if info.ModTime().After(p.destMTime) {
		p.destMTime = info.ModTime()
	}
}

func (p *Asset) performCopy() {
	log.Println("Perform Copy")

	if p.isFileAsset {
		sourcePath := p.sourcePaths[0]
		log.Println("destPath: ", p.destPath)

		info, _ := os.Stat(sourcePath)

		p.performCopyFile(p.destPath, sourcePath, info)
	} else {
		filepath.Walk(p.sourceDir, func(sourcePath string, info os.FileInfo, _ error) error {
			if p.sourceDir == sourcePath {
				return nil
			}

			if !p.recursive && info.IsDir() {
				return filepath.SkipDir
			}

			destPath := path.Join(p.destDir, strings.TrimPrefix(sourcePath, p.sourceDir))
			log.Println("destPath: ", destPath)

			if info.IsDir() {
				dirErr := os.MkdirAll(destPath, os.ModeDir|os.ModePerm)
				if dirErr != nil {
					log.Fatal("Failed to create directory ", destPath, dirErr)
				}
			} else {
				p.performCopyFile(destPath, sourcePath, info)
			}

			return nil
		})
	}
}

func (p *Asset) Copy(h http.Handler) http.Handler {
	log.Printf("Copy: %+v", p)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.MatchesRequestURI(r.RequestURI) && p.NeedsUpdate() {
			p.performCopy()
		}

		h.ServeHTTP(w, r)
	})
}
