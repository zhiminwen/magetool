package sshkit

import (
	"fmt"

	"github.com/gosuri/uilive"
)

type IOProgress struct {
	TotalSize int64
	SumSoFar  int64

	prefixDoing string
	prefixDone  string

	Writer *uilive.Writer
}

func NewIOProgress(total int64, prefixDoing, prefixDone string) *IOProgress {
	p := &IOProgress{
		TotalSize: total,
		Writer:    uilive.New(),

		prefixDoing: prefixDoing,
		prefixDone:  prefixDone,
	}

	p.Writer.Start()

	return p
}

func (p *IOProgress) Write(data []byte) (int, error) {
	n := len(data)
	p.SumSoFar = p.SumSoFar + int64(n)
	fmt.Fprintf(p.Writer, "%s... %0.2f%%\n", p.prefixDoing, 100*float64(p.SumSoFar)/float64(p.TotalSize))

	return n, nil
}

func (p *IOProgress) FinalMessage() {
	fmt.Fprintf(p.Writer, "%s.\n", p.prefixDone)
	p.Writer.Stop()
}
