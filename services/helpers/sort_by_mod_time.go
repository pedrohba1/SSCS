package helpers

import "io/fs"

// ByModTime sorts files by their modification time
type ByModTime []fs.FileInfo

func (s ByModTime) Len() int {
	return len(s)
}
func (s ByModTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByModTime) Less(i, j int) bool {
	// Note that we use After to sort oldest first
	return s[i].ModTime().Before(s[j].ModTime())
}
