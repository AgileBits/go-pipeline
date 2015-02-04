package pipeline

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (p *Asset) performBrowserify() {
	p.MkDestDir()

	browserifyPath, browserifyErr := exec.LookPath("browserify")
	if browserifyErr != nil {
		browserifyPath, _ = filepath.Abs("./node_modules/bin/browserify")
	}

	if _, statErr := os.Stat(browserifyPath); os.IsNotExist(statErr) {
		log.Fatal("Failed to find browserify")
	}

	cmd := exec.Command(browserifyPath, p.sourcePaths[0], "-o", p.destPath)
	// cmd.Env = os.Environ()
	log.Printf("Running `%s %s`", cmd.Path, strings.Join(cmd.Args, " "))
	cmdOutput, cmdErr := cmd.CombinedOutput()

	if cmdErr != nil {
		log.Printf("Browserify failed (%+v):\n%s", cmdErr, string(cmdOutput))
	}
}

func (p *Asset) Browserify(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.MatchesRequestURI(r.RequestURI) && p.NeedsUpdate() {
			p.performBrowserify()
		}

		h.ServeHTTP(w, r)
	})
}
