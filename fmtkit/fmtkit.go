package fmtkit

import (
	"log"
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

func (bf *BasicFormatter) Header(cmd string, args ...string) {
	if len(args) > 0 {
		log.Printf("===== Executing Command: '%s' at %s =====\n", color.HiYellowString("%s %s", cmd, args), time.Now().Format("15:04:05"))
	} else {
		log.Printf("===== Executing Command: '%s' at %s =====\n", color.HiYellowString("%s", cmd), time.Now().Format("15:04:05"))
	}
}

func (bf *BasicFormatter) NormalLine(prefix, line string) {
	log.Printf("%s %s\n", color.WhiteString(prefix), color.GreenString(line))
}

func (bf *BasicFormatter) ErrorLine(prefix, line string) {
	log.Printf("%s %s\n", color.WhiteString(prefix), color.RedString(line))
}

func (bf *BasicFormatter) Footer(duration time.Duration, err error) {
	var statusString string
	red := color.New(color.FgHiRed).Add(color.Bold).SprintfFunc()
	green := color.New(color.FgHiGreen).Add(color.Bold).SprintfFunc()

	if err != nil {
		statusString = red("X")
		log.Printf("===== Error: %s =====\n", color.RedString("%v", err))
	} else {
		statusString = green("√")
	}

	log.Printf("===== Ended at: %s (Total: %s) (%s) =====\n", time.Now().Format("15:04:05"), color.HiYellowString(duration.String()), statusString)
}
