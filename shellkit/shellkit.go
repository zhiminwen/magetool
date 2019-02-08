package shellkit

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/magefile/mage/sh"
	"github.com/zhiminwen/magetool/fmtkit"
)

var myfmt fmtkit.Formatter

func init() {
	myfmt = &fmtkit.BasicFormatter{}
}

func SetFormatter(fmt fmtkit.Formatter) {
	myfmt = fmt
}

func display(reader io.Reader, wg *sync.WaitGroup, dispFn func(string)) {
	r := bufio.NewReader(reader)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		dispFn(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Failed to read from reader:", err)
	}

	wg.Done()
}

// Execute the command
func Execute(cmd string, args ...string) {
	var env map[string]string
	execute(env, cmd, args...)
}

//Execute as shell command
func ExecuteShell(cmd string) {
	var shell, arg string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		arg = "/c"
	} else {
		shell = "sh"
		arg = "-c"
	}

	args := []string{arg, fmt.Sprintf(`""%s""`, cmd)} //double quote to work??
	Execute(shell, args...)
}

func ExecuteWith(env map[string]string, cmd string, args ...string) {
	execute(env, cmd, args...)
}

func execute(env map[string]string, cmd string, args ...string) {
	myfmt.Header(cmd, args...)
	startTime := time.Now()

	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Printf("Failed to get stdout. %v", err)
		return
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		log.Printf("Failed to get stderr. %v", err)
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go display(stdout, &wg, func(line string) {
		myfmt.NormalLine("Output:", line)
	})
	wg.Add(1)
	go display(stderr, &wg, func(line string) {
		myfmt.ErrorLine("Error:", line)
	})

	err = c.Run()
	wg.Wait()

	duration := time.Since(startTime)
	myfmt.Footer(duration, err)

	if err != nil {
		log.Fatalf("Failed to execute. Exit...")
	}
}

//Capture the output as string
func Capture(cmd string, args ...string) string {
	stdout, err := sh.Output(cmd, args...)
	if err != nil {
		fmt.Printf("%s", color.RedString("Failed to run %s %v. Error:%v", cmd, args, err))
		os.Exit(1)
	}
	// fmt.Printf("output:%s", stdout)
	stdout = strings.TrimSpace(stdout)
	return stdout
}
