package backend

import (
	"fmt"
	
	"github.com/spf13/afero"
)

var fs afero.Fs = nil

// InitFS creates a safe filesystem rooted in the given storeDir.
func InitFS(storeDir string) {
	fs = afero.NewBasePathFs(afero.NewOsFs(), storeDir)
}

// OpenFile retrieves a handle to a file stored on the server, rooted in the database folder.
func OpenFile(name string) (afero.File, error) {
	if fs == nil {
		return nil, fmt.Errorf("Filesystem is nil. Filesystem must be initialized with backend.InitFS(storeDir string)")
	}

	return fs.Open(name)
	
}