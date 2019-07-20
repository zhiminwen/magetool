// +build mage

package main

import (
	"os"

	"github.com/zhiminwen/magetool/shellkit"
)

func init() {
	os.Setenv("MAGEFILE_VERBOSE", "true")
}

func TestShell() {
	shellkit.ExecuteShell("ls *go")
	shellkit.ExecuteShell(`ls " *go"`)

}
