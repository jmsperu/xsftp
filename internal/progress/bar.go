package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Bar displays a progress bar for file transfers.
type Bar struct {
	name    string
	total   int64
	current int64
	mu      sync.Mutex
	start   time.Time
	done    bool
}

// NewBar creates a progress bar.
func NewBar(name string, total int64) *Bar {
	return &Bar{
		name:  name,
		total: total,
		start: time.Now(),
	}
}

// Writer wraps an io.Writer to track progress.
type Writer struct {
	bar *Bar
	w   io.Writer
}

// NewWriter creates a progress-tracking writer.
func NewWriter(bar *Bar, w io.Writer) *Writer {
	return &Writer{bar: bar, w: w}
}

func (pw *Writer) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	pw.bar.Add(int64(n))
	return n, err
}

// Reader wraps an io.Reader to track progress.
type Reader struct {
	bar *Bar
	r   io.Reader
}

// NewReader creates a progress-tracking reader.
func NewReader(bar *Bar, r io.Reader) *Reader {
	return &Reader{bar: bar, r: r}
}

func (pr *Reader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.bar.Add(int64(n))
	return n, err
}

// Add adds n bytes to the progress.
func (b *Bar) Add(n int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current += n
	b.render()
}

// Finish marks the bar as complete.
func (b *Bar) Finish() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = b.total
	b.done = true
	b.render()
	fmt.Println()
}

func (b *Bar) render() {
	width := 40
	var pct float64
	if b.total > 0 {
		pct = float64(b.current) / float64(b.total)
	}
	if pct > 1 {
		pct = 1
	}

	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("=", filled)
	if filled < width {
		bar += ">"
		bar += strings.Repeat(" ", width-filled-1)
	}

	elapsed := time.Since(b.start).Seconds()
	speed := float64(0)
	if elapsed > 0 {
		speed = float64(b.current) / elapsed
	}

	name := b.name
	if len(name) > 30 {
		name = "..." + name[len(name)-27:]
	}

	fmt.Printf("\r%-30s [%s] %s/%s %s/s",
		name, bar,
		formatBytes(b.current), formatBytes(b.total),
		formatBytes(int64(speed)))
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
