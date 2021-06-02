package fmtkit

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/fatih/color"
)

type Formatter interface {
	Header(cmd string, args ...string)
	NormalLine(prefix, line string)
	ErrorLine(prefix, line string)
	Footer(duration time.Duration, err error)
}

type BasicFormatter struct{}

//taking care for window's environment
func printf(format string, args ...string) {
	ifaceArgs := []interface{}{}
	for _, v := range args {
		ifaceArgs = append(ifaceArgs, v)
	}

	if runtime.GOOS == "windows" {
		fmt.Fprintf(color.Output, format, ifaceArgs...)
	} else {
		log.Printf(format, ifaceArgs...)
	}
}

func (bf *BasicFormatter) Header(cmd string, args ...string) {
	if len(args) > 0 {
		printf("===== Executing Command: '%s' at %s =====\n", color.HiYellowString("%s %s", cmd, args), time.Now().Format("15:04:05"))
	} else {
		printf("===== Executing Command: '%s' at %s =====\n", color.HiYellowString("%s", cmd), time.Now().Format("15:04:05"))
	}
}

func (bf *BasicFormatter) NormalLine(prefix, line string) {
	// printf("%s %s\n", color.WhiteString(prefix), color.GreenString(line))
	printf("%s\n", color.GreenString(line))
}

func (bf *BasicFormatter) ErrorLine(prefix, line string) {
	// printf("%s %s\n", color.WhiteString(prefix), color.RedString(line))
	printf("%s\n", color.RedString(line))
}

func (bf *BasicFormatter) Footer(duration time.Duration, err error) {
	var statusString string
	red := color.New(color.FgHiRed).Add(color.Bold).SprintfFunc()
	green := color.New(color.FgHiGreen).Add(color.Bold).SprintfFunc()

	if err != nil {
		statusString = red("X")
		printf("===== Error: %s =====\n", color.RedString("%v", err))
	} else {
		statusString = green("√")
	}

	printf("===== Ended at: %s (Total: %s) (%s) =====\n", time.Now().Format("15:04:05"), color.HiYellowString(duration.String()), statusString)
}
