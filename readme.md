# A Capistrano/SSHKit Implementation with Golang

## Download and Import
```
go get github.com/zhiminwen/magetool
```

```
import(
	"github.com/zhiminwen/magetool/shellkit"
	"github.com/zhiminwen/magetool/sshkit"
)
```

## Medium Story

## Synopsis

### Mage
Use [Mage](https://github.com/magefile/mage) to organize the tasks together

```golang
// +build mage
package main
import (
 "os"
 "github.com/zhiminwen/magetool/shellkit"
)
func init() {
 os.Setenv("MAGEFILE_VERBOSE", "true")
}
//Run Cmd
func RunCmd() {
 shellkit.Execute("cmd", "/c", "type test.log & sleep 10")
}
//dir
func Dir() {
 shellkit.Execute("cmd", "/c", "time /t 1>&2 & exit 1")
}
```

### Shellkit

- Use Shellkit.Execute() to run local command.
- Use Shellkit.Capture to capture command output.

### SSHkit

Init the client

```golang
//Use password
client1 := sshkit.NewSSHClient("192.168.5.10", "22", "ubuntu", "password", "")
//Use private key
s2 := sshkit.NewSSHClient("192.168.5.10", "22", "ubuntu", "", "mykeyfile"
// Use a slice of the client

var servers []*sshkit.SSHClient
servers = append(servers, sshkit.NewSSHClient("192.168.5.10", "22", "ubuntu", "password", ""))
servers = append(servers, sshkit.NewSSHClient("192.168.5.10", "22", "ubuntu2", "password", ""))
```

Execute a command

```golang
// Run against a single client
client1.Execute("id; hostname")
//Run against slice of servers
sshkit.Execute(servers, "id; hostname")
```

Capture a result

``` golang
result, err := client.Capture("hostname")
```

```golang
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
```


Upload and Download

```golang
client1.Upload("example.txt", "/tmp/example.txt")
client1.Download("/etc/hosts", "hosts.txt")
```

```
client1.Put(`This is a test content`, "mytest.txt")
list, err := servers[0].Get("/etc/passwd")
```

Interact with command

```golang
client1.Execute(`rm -rf .ssh/known_hosts`)
client1.ExecuteInteractively("scp ubuntu@ubuntu:/tmp/test.txt /tmp/test.txt.dupped", map[string]string{
  `\(yes/no\)`: "yes",
  "password":   "password",
 })
```

