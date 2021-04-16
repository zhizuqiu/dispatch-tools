package service

import (
	"github.com/cheggaaa/pb/v3"
	"io"
)

type Reader struct {
	io.Reader
	Current int64
	Bar     *pb.ProgressBar
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)

	r.Current += int64(n)
	r.Bar.SetCurrent(r.Current)
	return
}
