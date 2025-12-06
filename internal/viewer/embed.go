package viewer

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed ui/dist/*
var uiFS embed.FS

// UIHandler returns an http.Handler that serves the embedded UI files.
func UIHandler() http.Handler {
	// Strip the "ui/dist" prefix so files are served from root
	subFS, err := fs.Sub(uiFS, "ui/dist")
	if err != nil {
		panic("failed to create sub filesystem: " + err.Error())
	}
	return http.FileServer(http.FS(subFS))
}
