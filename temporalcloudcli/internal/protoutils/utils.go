package protoutils

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ClearDeprecatedFields(msg proto.Message) {
	clearDeprecatedFields(msg.ProtoReflect())
}

// StripDeprecatedJSONFields removes deprecated fields from a marshaled-protojson
// byte slice by walking the descriptor of msg in parallel with the parsed JSON.
// A field is considered deprecated if its proto name ends with "_deprecated" or
// its [deprecated = true] option is set.
func StripDeprecatedJSONFields(data []byte, msg proto.Message) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("unable to parse json for deprecated field removal: %v", err)
	}
	stripDeprecatedJSON(v, msg.ProtoReflect().Descriptor())
	out, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("unable to re-marshal json after deprecated field removal: %v", err)
	}
	return out, nil
}

func isDeprecatedField(fd protoreflect.FieldDescriptor) bool {
	if strings.HasSuffix(string(fd.Name()), "_deprecated") {
		return true
	}
	opts, _ := fd.Options().(*descriptorpb.FieldOptions)
	return opts.GetDeprecated()
}

func stripDeprecatedJSON(v any, md protoreflect.MessageDescriptor) {
	obj, ok := v.(map[string]any)
	if !ok {
		return
	}
	fields := md.Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		key := fd.JSONName()
		if isDeprecatedField(fd) {
			delete(obj, key)
			continue
		}
		if fd.Kind() != protoreflect.MessageKind && fd.Kind() != protoreflect.GroupKind {
			continue
		}
		child, present := obj[key]
		if !present {
			continue
		}
		switch {
		case fd.IsMap():
			if fd.MapValue().Kind() != protoreflect.MessageKind {
				continue
			}
			entries, ok := child.(map[string]any)
			if !ok {
				continue
			}
			for _, entry := range entries {
				stripDeprecatedJSON(entry, fd.MapValue().Message())
			}
		case fd.IsList():
			items, ok := child.([]any)
			if !ok {
				continue
			}
			for _, item := range items {
				stripDeprecatedJSON(item, fd.Message())
			}
		default:
			stripDeprecatedJSON(child, fd.Message())
		}
	}
}

// clearDeprecatedFields recursively clears fields whose proto name ends with
// "_deprecated" or that are marked with [deprecated = true] in the proto file,
// and recurses into nested messages.
//
// AIDEV-NOTE: Fields are collected before clearing to avoid mutating the message
// during Range iteration, which has undefined behavior in the protobuf-go runtime.
func clearDeprecatedFields(msg protoreflect.Message) {
	var toClear []protoreflect.FieldDescriptor

	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if isDeprecatedField(fd) {
			toClear = append(toClear, fd)
			return true
		}
		switch fd.Kind() {
		case protoreflect.MessageKind, protoreflect.GroupKind:
			if fd.IsList() {
				list := v.List()
				for i := range list.Len() {
					clearDeprecatedFields(list.Get(i).Message())
				}
			} else if fd.IsMap() {
				if fd.MapValue().Kind() == protoreflect.MessageKind {
					v.Map().Range(func(_ protoreflect.MapKey, mv protoreflect.Value) bool {
						clearDeprecatedFields(mv.Message())
						return true
					})
				}
			} else {
				clearDeprecatedFields(v.Message())
			}
		}
		return true
	})

	for _, fd := range toClear {
		msg.Clear(fd)
	}
}
