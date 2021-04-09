package tests

import (
	"github.com/cheggaaa/pb/v3"
	"testing"
	"time"
)

func TestPb(t *testing.T) {
	count := 2000
	// create and start new bar
	// bar := pb.StartNew(count)

	tmpl := `{{ green "downloading:" }} {{ bar . "[" "-" (cycle . "↖" "↗" "↘" "↙" ) "." "]"}} {{percent .}} {{counters . }} `
	// start bar based on our template
	bar := pb.ProgressBarTemplate(tmpl).Start64(int64(count))

	// start bar from 'default' template
	// bar := pb.Default.Start(count)

	// start bar from 'simple' template
	// bar := pb.Simple.Start(count)

	// start bar from 'full' template
	// bar := pb.Full.Start(count)

	for i := 0; i < count; i++ {
		// bar.Increment()
		bar.SetCurrent(int64(i + 1))
		time.Sleep(time.Millisecond)
	}
	bar.Finish()
}
