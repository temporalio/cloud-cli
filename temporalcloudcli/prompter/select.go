package prompter

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"golang.org/x/term"
)

// ErrSelectCanceled is returned when the user cancels a selection (Ctrl-C, Ctrl-D or q).
var ErrSelectCanceled = errors.New("selection canceled")

// Input bytes. Raw mode delivers keys as bytes instead of acting on them.
const (
	keyCtrlC = '\x03'
	keyCtrlD = '\x04'
	keyEsc   = '\x1b'

	// Arrow keys arrive as ESC, a prefix, then a direction: ESC [ A, or ESC O A
	// when the terminal is in application cursor mode.
	arrowPrefixCSI = '['
	arrowPrefixSS3 = 'O'
	arrowUp        = 'A'
	arrowDown      = 'B'
)

// Output escape sequences (ANSI/VT100).
const (
	hideCursor = "\x1b[?25l"
	showCursor = "\x1b[?25h"
	cursorUp   = "\x1b[1A" // move the cursor up one line
	clearLine  = "\x1b[2K" // erase the line the cursor is on
	// Raw mode disables output post-processing, so lines must end in \r\n.
	newline = "\r\n"
)

// fdFile is satisfied by *os.File, the only kind of stream that can be a terminal.
type fdFile interface {
	Fd() uintptr
}

// IsTerminal reports whether in and out are both interactive terminals, so an
// arrow-key prompt can be shown.
func IsTerminal(in io.Reader, out io.Writer) bool {
	return isTerminal(in) && isTerminal(out)
}

func isTerminal(stream any) bool {
	f, ok := stream.(fdFile)
	return ok && term.IsTerminal(int(f.Fd()))
}

// SelectInTerminal puts in into raw mode, shows an arrow-key menu on out and
// returns the index of the chosen option. The caller must check IsTerminal
// first.
func SelectInTerminal(in io.Reader, out io.Writer, message string, options []string) (int, error) {
	f, ok := in.(fdFile)
	if !ok {
		return 0, errors.New("input is not a terminal")
	}
	fd := int(f.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, fmt.Errorf("failed to read from terminal: %w", err)
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	return Select(in, out, message, options)
}

// Select runs an arrow-key menu over in and out, which must already be in raw
// mode. Up/Down (or k/j) move, Enter chooses, and Ctrl-C, Ctrl-D or q cancel.
// It returns the index of the chosen option.
func Select(in io.Reader, out io.Writer, message string, options []string) (int, error) {
	if len(options) == 0 {
		return 0, errors.New("no options to select from")
	}

	r := bufio.NewReader(in)
	selected := 0

	fmt.Fprint(out, hideCursor, message, newline)
	defer fmt.Fprint(out, showCursor)
	drawOptions(out, options, selected)

	for {
		key, err := readKey(r)
		if err != nil {
			return 0, err
		}
		switch key {
		case keyUp:
			selected = (selected - 1 + len(options)) % len(options)
		case keyDown:
			selected = (selected + 1) % len(options)
		case keyEnter:
			clearOptions(out, len(options))
			fmt.Fprint(out, cursorUp, clearLine, message, " ", options[selected], newline)
			return selected, nil
		case keyCancel:
			clearOptions(out, len(options))
			return 0, ErrSelectCanceled
		default:
			continue
		}
		clearOptions(out, len(options))
		drawOptions(out, options, selected)
	}
}

type key int

const (
	keyOther key = iota
	keyUp
	keyDown
	keyEnter
	keyCancel
)

func readKey(r *bufio.Reader) (key, error) {
	b, err := r.ReadByte()
	if errors.Is(err, io.EOF) {
		return keyCancel, nil
	}
	if err != nil {
		return keyOther, fmt.Errorf("failed to read from terminal: %w", err)
	}
	switch b {
	case '\r', '\n':
		return keyEnter, nil
	case keyCtrlC, keyCtrlD, 'q':
		return keyCancel, nil
	case 'k':
		return keyUp, nil
	case 'j':
		return keyDown, nil
	case keyEsc:
		prefix, err := r.ReadByte()
		if err != nil || (prefix != arrowPrefixCSI && prefix != arrowPrefixSS3) {
			return keyOther, nil
		}
		code, err := r.ReadByte()
		if err != nil {
			return keyOther, nil
		}
		switch code {
		case arrowUp:
			return keyUp, nil
		case arrowDown:
			return keyDown, nil
		}
	}
	return keyOther, nil
}

func drawOptions(out io.Writer, options []string, selected int) {
	for i, opt := range options {
		marker := "  "
		if i == selected {
			marker = "> "
		}
		fmt.Fprint(out, marker, opt, newline)
	}
}

// clearOptions moves the cursor back over the drawn options and erases them.
func clearOptions(out io.Writer, n int) {
	for range n {
		fmt.Fprint(out, cursorUp, clearLine)
	}
}
