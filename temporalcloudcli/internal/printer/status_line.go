package printer

import (
	"io"
	"os"

	"golang.org/x/term"
)

// clearLine erases the current terminal line and returns the cursor to its start.
const clearLine = "\033[2K\r"

var ellipsisFrames = []string{"", ".", "..", "..."}

// StatusLine reports the progress of a long-running operation. On a terminal it rewrites one line
// in place with an animated ellipsis. Anywhere else it writes one line per update, so piped and
// captured output is unaffected. Like Print, it writes nothing in JSON mode.
type StatusLine struct {
	printer  *Printer
	terminal bool
	frame    int
	open     bool
}

// NewStatusLine returns a StatusLine that writes through p.
func (p *Printer) NewStatusLine() *StatusLine {
	return &StatusLine{printer: p, terminal: isTerminal(p.Output)}
}

// Progress reports an intermediate state. msg must not end in an ellipsis; one is added.
func (s *StatusLine) Progress(msg string) {
	if !s.terminal {
		s.printer.Print(msg + "...\n")
		return
	}
	s.printer.Print(clearLine + msg + ellipsisFrames[s.frame])
	s.frame = (s.frame + 1) % len(ellipsisFrames)
	s.open = true
}

// Done reports the final state and leaves it on its own line.
func (s *StatusLine) Done(msg string) {
	if s.open {
		msg = clearLine + msg
	}
	s.printer.Print(msg + "\n")
	s.open = false
}

// Clear erases an in-progress line so that whatever prints next starts on a clean line.
func (s *StatusLine) Clear() {
	if s.open {
		s.printer.Print(clearLine)
		s.open = false
	}
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}
