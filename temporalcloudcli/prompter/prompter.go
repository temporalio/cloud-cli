package prompter

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type Prompter interface {
	PromptApply(old, new proto.Message, verbose bool) error
	PromptYes(message string, autoConfirm bool) (bool, error)
}

type prompter struct {
	printer     *printer.Printer
	stdin       io.Reader
	autoConfirm bool
}

func NewPrompter(p *printer.Printer, si io.Reader, autoConfirm bool) Prompter {
	return &prompter{
		printer:     p,
		stdin:       si,
		autoConfirm: autoConfirm,
	}
}

func (p *prompter) PromptApply(existing, modified proto.Message, verbose bool) error {
	if p.printer.JSON {
		p.printer.PrintDiff(existing, modified, printer.DiffOptions{
			Verbose: verbose,
		})
	}

	yes, err := p.PromptYes("Apply (y/yes)?", p.autoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("Aborting apply.")
	}
	return nil
}

func (p *prompter) PromptYes(message string, autoConfirm bool) (bool, error) {
	if p.printer.JSON && !autoConfirm {
		return false, fmt.Errorf("must bypass prompts when using JSON output")
	}
	p.printer.Print(message, " ")
	if autoConfirm {
		p.printer.Println("yes")
		return true, nil
	}
	line, _ := bufio.NewReader(p.stdin).ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}
