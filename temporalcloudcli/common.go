package temporalcloudcli

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func isNothingChangedErr(idempotent bool, e error) bool {
	// If we are not idempotent, we should error on nothing to change
	if !idempotent {
		return false
	}

	s, ok := status.FromError(e)
	if !ok {
		return false
	}
	return s.Code() == codes.InvalidArgument && strings.Contains(s.Message(), "nothing to change")
}

// loadJSONSpec loads a JSON specification from either a file path (prefixed with '@')
// or treats the input as inline JSON. Returns the parsed data as a byte slice.
func loadJSONSpec(spec string) ([]byte, error) {
	// Check if spec starts with '@' indicating file path
	if strings.HasPrefix(spec, "@") {
		// Remove '@' prefix and read file
		filePath := strings.TrimPrefix(spec, "@")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read spec file %q: %w", filePath, err)
		}
		return data, nil
	}

	// Treat as inline JSON
	return []byte(spec), nil
}

func runEditorForJSONEdit(existing, valuePtr any) error {
	existingBytes, err := json.MarshalIndent(existing, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to convert existing object to json: %v", err)
	}
	updatedBytes, err := runEditor(existingBytes)
	if err != nil {
		return err
	}
	return json.Unmarshal(updatedBytes, valuePtr)
}

func runEditorForJSONEditForProtos(existing, value proto.Message) error {
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		Indent:          "    ",
	}
	existingBytes, err := marshaler.Marshal(existing)
	if err != nil {
		return fmt.Errorf("unable to convert existing object to json: %v", err)
	}
	updatedBytes, err := runEditor(existingBytes)
	if err != nil {
		return err
	}
	unmarshaller := protojson.UnmarshalOptions{}
	return unmarshaller.Unmarshal(updatedBytes, value)
}

func runEditor(existing []byte) ([]byte, error) {
	f, err := os.CreateTemp("", "cloud-cli-edit-*.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file for editing: %v", err)
	}

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
		return nil, fmt.Errorf("error executing %q: %v", editor, err)
	}

	updated, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to read updated data from temp file: %v", err)
	}

	if bytes.Equal(existing, updated) {
		return nil, fmt.Errorf("no changes detected")
	}
	return updated, nil
}
