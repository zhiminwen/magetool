package shellkit

import "testing"

func TestExecute(t *testing.T) {
	Execute("cmd", "/c", "dir")
}
