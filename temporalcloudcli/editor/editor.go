package editor

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type (
	Editor interface {
		EditProto(existing proto.Message) (proto.Message, error)
	}

	editor struct{}
)

func NewEditor() Editor {
	return &editor{}
}

func (e *editor) EditProto(existing proto.Message) (proto.Message, error) {
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		Indent:          "    ",
	}
	existingBytes, err := marshaler.Marshal(existing)
	if err != nil {
		return nil, fmt.Errorf("unable to convert existing object to json: %v", err)
	}
	updatedBytes, err := e.runEditor(existingBytes)
	if err != nil {
		return nil, err
	}
	value := existing.ProtoReflect().New().Interface()
	if err := protojson.Unmarshal(updatedBytes, value); err != nil {
		return nil, fmt.Errorf("unable to convert updated json to object: %v", err)
	}
	if proto.Equal(existing, value) {
		return nil, fmt.Errorf("no changes detected")
	}
	return value, nil
}

func (e *editor) runEditor(existing []byte) ([]byte, error) {
	f, err := os.CreateTemp("", "cloud-cli-edit-*.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file for editing: %v", err)
	}

	defer func() {
		// Clean up temp file.
		f.Close()
		_ = os.Remove(f.Name())
	}()

	if _, err := f.Write(existing); err != nil {
		return nil, fmt.Errorf("unable to write existing data to temp file for editing: %v", err)
	}

	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("unable to close temp file: %v", err)
	}

	editor := strings.Split(cmp.Or(os.Getenv("VISUAL"), os.Getenv("EDITOR"), "vim"), " ")
	program, args := editor[0], editor[1:]

	cmd := exec.Command(program, append(args, f.Name())...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing %q: %v", strings.Join(editor, " "), err)
	}

	updated, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to read updated data from temp file: %v", err)
	}

	return updated, nil
}
