// Copyright (c) 2023 Multus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmdutils is the package that contains utilities for multus command
package cmdutils

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFileAtomic does file copy atomically
func CopyFileAtomic(srcFilePath, destDir, tempFileName, destFileName string) error {
	destFilePath := filepath.Join(destDir, destFileName)
	destMD5, _ := getMD5FromFile(destFilePath)
	if destMD5 != nil {
		srcMD5, _ := getMD5FromFile(srcFilePath)
		if bytes.Compare(destMD5, srcMD5) == 0 {
			fmt.Fprintf(os.Stderr, "not need to copy file, %s and %s is same\n", srcFilePath, destFilePath)
			return nil
		}
	}
	tempFilePath := filepath.Join(destDir, tempFileName)
	// check temp filepath and remove old file if exists
	if _, err := os.Stat(tempFilePath); err == nil {
		err = os.Remove(tempFilePath)
		if err != nil {
			return fmt.Errorf("cannot remove old temp file %q: %v", tempFilePath, err)
		}
	}

	// create temp file
	f, err := os.CreateTemp(destDir, tempFileName)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("cannot create temp file %q in %q: %v", tempFileName, destDir, err)
	}

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return fmt.Errorf("cannot open file %q: %v", srcFilePath, err)
	}
	defer srcFile.Close()

	// Copy file to tempfile
	_, err = io.Copy(f, srcFile)
	if err != nil {
		f.Close()
		os.Remove(tempFilePath)
		return fmt.Errorf("cannot write data to temp file %q: %v", tempFilePath, err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("cannot flush temp file %q: %v", tempFilePath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("cannot close temp file %q: %v", tempFilePath, err)
	}

	// change file mode if different
	// destFilePath := filepath.Join(destDir, destFileName)
	_, err = os.Stat(destFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	srcFileStat, err := os.Stat(srcFilePath)
	if err != nil {
		return err
	}

	if err := os.Chmod(f.Name(), srcFileStat.Mode()); err != nil {
		return fmt.Errorf("cannot set stat on temp file %q: %v", f.Name(), err)
	}

	// replace file with tempfile
	if err := os.Rename(f.Name(), destFilePath); err != nil {
		return fmt.Errorf("cannot replace %q with temp file %q: %v", destFilePath, tempFilePath, err)
	}

	return nil
}

func getMD5FromFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
