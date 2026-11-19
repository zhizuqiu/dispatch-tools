package service

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cheggaaa/pb/v3"
)

func renderBar(t *testing.T, tmpl string, total, current int64) string {
	t.Helper()
	var buf bytes.Buffer
	bar := pb.ProgressBarTemplate(tmpl).New(0)
	bar.SetWriter(&buf)
	bar.SetTotal(total)
	bar.Start()
	bar.SetCurrent(current)
	bar.Finish()
	return buf.String()
}

func TestProgressBarShowsEmptyFileForZeroSize(t *testing.T) {
	templates := map[string]string{
		"upload":   uploadBarTmpl,
		"download": downloadBarTmpl,
	}
	for name, tmpl := range templates {
		t.Run(name, func(t *testing.T) {
			out := renderBar(t, tmpl, 0, 0)
			if !strings.Contains(out, "空文件") {
				t.Fatalf("output = %q, want 空文件 hint", out)
			}
			if strings.Contains(out, "?%") {
				t.Fatalf("output = %q, must not contain ?%%", out)
			}
		})
	}
}

func TestProgressBarRendersPercentForNonEmptyFile(t *testing.T) {
	templates := map[string]string{
		"upload":   uploadBarTmpl,
		"download": downloadBarTmpl,
	}
	for name, tmpl := range templates {
		t.Run(name, func(t *testing.T) {
			out := renderBar(t, tmpl, 10, 10)
			if strings.Contains(out, "空文件") {
				t.Fatalf("output = %q, must not show 空文件 for non-empty file", out)
			}
			if !strings.Contains(out, "100") {
				t.Fatalf("output = %q, want 100%%", out)
			}
		})
	}
}
