package prompter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// ErrSelectCanceled is returned when the user cancels a selection (Ctrl-C,
// Ctrl-D or q).
var ErrSelectCanceled = errors.New("selection canceled")

const (
	keyCtrlC = 3
	keyCtrlD = 4
	keyEsc   = 27
)

// IsTerminal reports whether stdin and stdout are both interactive terminals,
// so an arrow-key prompt can be shown.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// SelectInTerminal shows an arrow-key menu on stdout and returns the index of
// the chosen option. The caller must check IsTerminal first.
func SelectInTerminal(message string, options []string) (int, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, fmt.Errorf("failed to read from terminal: %w", err)
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	return Select(os.Stdin, os.Stdout, message, options)
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

	// Raw mode disables output post-processing, so lines end in \r\n.
	fmt.Fprintf(out, "\x1b[?25l%s\r\n", message)
	defer fmt.Fprint(out, "\x1b[?25h")
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
			fmt.Fprintf(out, "\x1b[1A\x1b[2K%s %s\r\n", message, options[selected])
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
		// Arrow keys arrive as ESC [ A/B, or ESC O A/B in application cursor mode.
		prefix, err := r.ReadByte()
		if err != nil || (prefix != '[' && prefix != 'O') {
			return keyOther, nil
		}
		code, err := r.ReadByte()
		if err != nil {
			return keyOther, nil
		}
		switch code {
		case 'A':
			return keyUp, nil
		case 'B':
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
		fmt.Fprintf(out, "%s%s\r\n", marker, opt)
	}
}

// clearOptions moves the cursor back over the drawn options and erases them.
func clearOptions(out io.Writer, n int) {
	for range n {
		fmt.Fprint(out, "\x1b[1A\x1b[2K")
	}
}
