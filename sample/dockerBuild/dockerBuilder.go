// +build mage

package main

import (
	"os"

	"github.com/zhiminwen/magetool/shellkit"
	"github.com/zhiminwen/magetool/sshkit"
	"github.com/zhiminwen/quote"
)

var (
	master    *sshkit.SSHClient
	targetDir string
)

func init() {
	os.Setenv("MAGEFILE_VERBOSE", "true")
	master = sshkit.NewSSHClient("192.168.xxx.yyy", "22", "ubuntu", "", "mykeyfile")
	targetDir = "cicdtools"
}

//Cross compile for Linux
func T01_CrossCompile() {
	cmd := quote.Cmd(`
		cd cicdtools
		set GOARCH=amd64
		set GOOS=linux
		mage -compile cicdbuilder
	`, "&&")
	shellkit.Execute("cmd", "/c", cmd)
}

//Upload cross compiled exe
func T02_Upload() {
	master.Execute("mkdir -p " + targetDir)
	master.Upload("cicdtools/cicdbuilder", targetDir+"/cicdbuilder")
	master.Execute("chmod a+rx " + targetDir + "/cicdbuilder")
}

//Build docker image
func T03_Dockerfile() {
	block := quote.Template(`
		apk add --update ca-certificates curl tar

		cd /tmp
		curl -LO https://github.com/magefile/mage/releases/download/v1.7.1/{{ .magetar }}
		tar zxvf {{ .magetar }}
		mv mage /usr/local/bin

		curl -LO https://download.docker.com/linux/static/stable/x86_64/{{ .dockertar }}
		tar zxvf {{ .dockertar }}
		mv docker/docker /usr/local/bin

		curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
		mv kubectl /usr/local/bin/kubectl
		chmod a+rx /usr/local/bin/kubectl

		rm -rf /tmp/*
	`, map[string]string{
		"magetar":   "mage_1.7.1_Linux-64bit.tar.gz",
		"dockertar": "docker-18.06.1-ce.tgz",
	})

	content := quote.HereDocf(`
		FROM alpine
		RUN %s
		ADD cicdbuilder /usr/local/bin

		CMD ["cat"]
	`, quote.Cmd(block, " && "))

	master.Execute("mkdir -p " + targetDir)
	master.Put(content, targetDir+"/Dockerfile")

	master.Execute(quote.Cmdf(`
		cd %s
		sudo docker build -t zhiminwen/cicdtool .
	`, " && ", targetDir))

}

func T04_Push() {
	lines := quote.Template(`
		sudo docker login -u zhiminwen -p {{ .DockerhubPass }}
		sudo docker push zhiminwen/cicdtool:latest
	`, map[string]string{
		"DockerhubPass": os.Getenv("MY_DOCKER_CREDENTIAL"),
	})

	master.Execute(quote.Cmd(lines, " && "))
}
