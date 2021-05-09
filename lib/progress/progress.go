package progress

import (
	"github.com/schollz/progressbar/v2"
)

type Bar interface {
	Add(val int)
	Finish()
}

var Prefix = ""

type bar struct {
	pb *progressbar.ProgressBar
}

func Start(descr string, size int) Bar {
	pb := progressbar.NewOptions(size, progressbar.OptionSetRenderBlankState(true), progressbar.OptionSetDescription(Prefix+descr))
	pb.RenderBlank()
	return &bar{pb}
}

func (b *bar) Add(val int) {
	b.pb.Add(val)
}

func (b *bar) Finish() {
	b.pb.Clear()
}
