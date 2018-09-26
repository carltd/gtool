package utils

import (
	"os"
	"strings"
	"io/ioutil"
	"path/filepath"
)

type ReplaceHelper struct {
	Root    string // The root directory
	OldText string // Replace with old text
	NewText string // Replace with new text
}

func (h *ReplaceHelper) DoWrok() error {

	return filepath.Walk(h.Root, h.walkCallback)

}

func (h ReplaceHelper) walkCallback(path string, f os.FileInfo, err error) error {

	if err != nil {
		return err
	}
	if f == nil {
		return nil
	}
	if f.IsDir() {
		return nil
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(buf)

	// To replace
	newContent := strings.Replace(content, h.OldText, h.NewText, -1)

	// To write
	ioutil.WriteFile(path, []byte(newContent), 0)

	return err
}
