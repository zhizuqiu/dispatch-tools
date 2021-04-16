package service

import (
	"github.com/cheggaaa/pb/v3"
	"io"
)

type Writer struct {
	io.Writer
	Total   int64
	Current int64
	Bar     *pb.ProgressBar
}

func (r *Writer) Write(p []byte) (n int, err error) {
	n, err = r.Writer.Write(p)

	r.Current += int64(n)
	r.Bar.SetCurrent(r.Current)
	return
}
