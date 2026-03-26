package prompter

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type Prompter interface {
	PromptApply(existing, modified proto.Message, verbose bool) (bool, error)
	PromptYes(message string) (bool, error)
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

func (p *prompter) PromptApply(existing, modified proto.Message, verbose bool) (bool, error) {
	if !p.printer.JSON {
		p.printer.PrintDiff(existing, modified, printer.DiffOptions{
			Verbose: verbose,
		})
	}

	yes, err := p.PromptYes("Apply")
	if err != nil {
		return false, err
	}
	if !yes {
		p.printer.Println("Aborting apply.")
	}
	return yes, nil
}

func (p *prompter) PromptYes(message string) (bool, error) {
	if p.printer.JSON && !p.autoConfirm {
		return false, fmt.Errorf("must bypass prompts when using JSON output")
	}
	p.printer.Print(message, " (y/yes)? ")
	if p.autoConfirm {
		p.printer.Println("yes")
		return true, nil
	}
	line, _ := bufio.NewReader(p.stdin).ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}
