// +build mage

package main

import (
	"fmt"
	"log"

	"github.com/zhiminwen/magetool/sshkit"
)

var servers []*sshkit.SSHClient

func init() {
	servers = append(servers, sshkit.NewSSHClient("192.168.5.10", "22", "ubuntu", "password", ""))
	servers = append(servers, sshkit.NewSSHClient("192.168.5.10", "22", "ubuntu2", "password", ""))
}

//Capture Hostname
func SSHGetHostname() {
	result, err := sshkit.Capture(servers[0], "hostname123")

	if err != nil {
		log.Fatalf("failed to capture")
	}

	fmt.Printf("result=%s", result)
}

// Execute command
func SSHExecute() {
	sshkit.Execute(servers, "id; hostname")
}

// Fail
func SSHFailure() {
	sshkit.Execute(servers, "id; hostname; unknown")
}

// Execute Func Block
func RunBlock() {
	sshkit.ExecuteFunc(servers, func(t *sshkit.SSHClient) {
		id, err := t.Capture("id -u")
		if err != nil {
			log.Printf("Failed to get id:%v", err)
			return
		}

		if id == "1000" {
			return
		}

		cmd := fmt.Sprintf("echo %s; hostname", id)
		t.Execute(cmd)
	})
}

// Execute interactively
func Interact() {
	servers[0].Execute(`rm -rf .ssh/known_hosts`)
	servers[0].ExecuteInteractively("scp ubuntu@ubuntu:/tmp/test.txt /tmp/test.txt.dupped", map[string]string{
		`\(yes/no\)`: "yes",
		"password":   "password",
	})
}

// Upload
func Upload() {
	servers[0].Upload(`C:\Users\IBM_ADMIN\Downloads\CK1051course.pdf`, "/tmp/sshtask.go")
}

// Upload string
func Put() {
	servers[0].Put(`This is a test string`, "/tmp/test.txt")
}

// put string to default home dir
func PutToHome() {
	servers[0].Put(`This is a test string`, "test.txt")
}

// Download
func Download() {
	servers[0].Download("/etc/passwd.nonexist.file", "passwd.txt")

	// servers[0].Download("/usr/local/bin/kubectl", "kubectl")
}

// test Get
func Get() {
	list, err := servers[0].Get("/etc/passwd")
	if err != nil {
		log.Printf("Failed to get file")
	}
	log.Printf("result: %s", list)
}
