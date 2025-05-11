package sshkit

import (

	// "github.com/gosuri/uilive"

	"github.com/schollz/progressbar/v3"
)

type IOProgress struct {
	TotalSize int64
	SumSoFar  int64

	prefixDoing string
	prefixDone  string

	// Writer *uilive.Writer
	Bar *progressbar.ProgressBar
}

func NewIOProgress(total int64, prefixDoing, prefixDone string) *IOProgress {
	p := &IOProgress{
		TotalSize: total,
		Bar:       progressbar.DefaultBytes(-1, prefixDoing),

		prefixDoing: prefixDoing,
		prefixDone:  prefixDone,
	}

	return p
}

// func (p *IOProgress) Write(data []byte) (int, error) {
// 	n := len(data)
// 	// p.SumSoFar = p.SumSoFar + int64(n)
// 	// fmt.Fprintf(p.Writer, "%s... %0.2f%%\n", p.prefixDoing, 100*float64(p.SumSoFar)/float64(p.TotalSize))

// 	return n, nil
// }

// func (p *IOProgress) FinalMessage() {
// 	// fmt.Fprintf(p.Writer, "%s.\n", p.prefixDone)
// 	// p.Writer.Stop()
// }
