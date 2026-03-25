// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.

package keystore

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ava-labs/libevm/log"
	mapset "github.com/deckarep/golang-set/v2"
)

// fileCache is a cache of files seen during scan of keystore.
type fileCache struct {
	all     mapset.Set[string]   // Set of all files from the keystore folder
	modTime map[string]time.Time // path -> mod time from the previous scan
	mu      sync.Mutex
}

// scan performs a new scan on the given directory, compares against the already
// cached filenames, and returns file sets: creates, deletes, updates.
func (fc *fileCache) scan(keyDir string) (mapset.Set[string], mapset.Set[string], mapset.Set[string], error) {
	t0 := time.Now()

	// List all the files from the keystore folder
	files, err := os.ReadDir(keyDir)
	if err != nil {
		return nil, nil, nil, err
	}
	t1 := time.Now()

	fc.mu.Lock()
	defer fc.mu.Unlock()

	if fc.modTime == nil {
		fc.modTime = make(map[string]time.Time)
	}

	// Iterate all the files and gather their metadata
	all := mapset.NewThreadUnsafeSet[string]()
	newModTimes := make(map[string]time.Time)

	for _, fi := range files {
		path := filepath.Join(keyDir, fi.Name())
		// Skip any non-key files from the folder
		if nonKeyFile(fi) {
			log.Trace("Ignoring file on account scan", "path", path)
			continue
		}
		all.Add(path)

		info, err := fi.Info()
		if err != nil {
			return nil, nil, nil, err
		}
		newModTimes[path] = info.ModTime()
	}
	t2 := time.Now()

	// Update the tracked files and return the three sets
	deletes := fc.all.Difference(all) // Deletes = previous - current
	creates := all.Difference(fc.all) // Creates = current - previous

	// Treat a path as updated if it already existed and its mod time changed.
	// Per-path comparison avoids missing writes after os.Chtimes moved the
	// timestamp backward, or when several files share the same coarse mtime.
	updates := mapset.NewThreadUnsafeSet[string]()
	for _, path := range all.ToSlice() {
		if creates.Contains(path) {
			continue
		}
		if !fc.all.Contains(path) {
			continue
		}
		old, had := fc.modTime[path]
		if !had || !old.Equal(newModTimes[path]) {
			updates.Add(path)
		}
	}

	fc.all = all
	fc.modTime = newModTimes
	t3 := time.Now()

	// Report on the scanning stats and return
	log.Debug("FS scan times", "list", t1.Sub(t0), "set", t2.Sub(t1), "diff", t3.Sub(t2))
	return creates, deletes, updates, nil
}

// nonKeyFile ignores editor backups, hidden files and folders/symlinks.
func nonKeyFile(fi os.DirEntry) bool {
	// Skip editor backups and UNIX-style hidden files.
	if strings.HasSuffix(fi.Name(), "~") || strings.HasPrefix(fi.Name(), ".") {
		return true
	}
	// Skip misc special files, directories (yes, symlinks too).
	if fi.IsDir() || !fi.Type().IsRegular() {
		return true
	}
	return false
}
