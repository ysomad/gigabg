package widget

import (
	"sync"

	"golang.design/x/clipboard"
)

var clipboardOnce sync.Once

func readClipboard() string {
	clipboardOnce.Do(func() {
		_ = clipboard.Init() //nolint:errcheck // best-effort clipboard init
	})
	return string(clipboard.Read(clipboard.FmtText))
}
