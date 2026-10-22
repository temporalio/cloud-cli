package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusLine_NotTerminal(t *testing.T) {
	var buf bytes.Buffer
	s := (&Printer{Output: &buf}).NewStatusLine()

	s.Progress("[10:00:00] Operation pending")
	s.Progress("[10:00:01] Operation in progress")
	s.Done("[10:00:02] Operation completed successfully")
	s.Clear()

	require.Equal(t,
		"[10:00:00] Operation pending...\n"+
			"[10:00:01] Operation in progress...\n"+
			"[10:00:02] Operation completed successfully\n",
		buf.String())
}

func TestStatusLine_TerminalRewritesOneLine(t *testing.T) {
	var buf bytes.Buffer
	s := &StatusLine{printer: &Printer{Output: &buf}, terminal: true}

	for range 5 {
		s.Progress("working")
	}
	s.Done("finished")

	require.Equal(t,
		clearLine+"working"+
			clearLine+"working."+
			clearLine+"working.."+
			clearLine+"working..."+
			clearLine+"working"+
			clearLine+"finished\n",
		buf.String())
}

func TestStatusLine_TerminalDoneWithoutProgress(t *testing.T) {
	var buf bytes.Buffer
	s := &StatusLine{printer: &Printer{Output: &buf}, terminal: true}

	s.Done("finished")

	require.Equal(t, "finished\n", buf.String())
}

func TestStatusLine_TerminalClearOnlyErasesAnOpenLine(t *testing.T) {
	var buf bytes.Buffer
	s := &StatusLine{printer: &Printer{Output: &buf}, terminal: true}

	s.Clear()
	s.Progress("working")
	s.Clear()
	s.Clear()

	require.Equal(t, clearLine+"working"+clearLine, buf.String())
}

func TestStatusLine_JSONWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	s := &StatusLine{printer: &Printer{Output: &buf, JSON: true}, terminal: true}

	s.Progress("working")
	s.Done("finished")
	s.Clear()

	require.Empty(t, buf.String())
}
