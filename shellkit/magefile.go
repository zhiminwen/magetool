// +build mage

package main

import (
	"github.com/zhiminwen/magetool/shellkit"
)

// type NS mg.Namespace

//Run Cmd
func RunCmd() {
	shellkit.Execute("cmd", "/c", "type test.log & sleep 10 & ")
}

//dir
func Dir() {
	shellkit.Execute("cmd", "/c", "dir 1>&2")
}
