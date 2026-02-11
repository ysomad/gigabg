package widget

import "golang.design/x/clipboard"

func init() {
	clipboard.Init()
}

func readClipboard() string {
	return string(clipboard.Read(clipboard.FmtText))
}
