package protoutils

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ClearDeprecatedFields(msg proto.Message) {
	clearDeprecatedFields(msg.ProtoReflect())
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
		opts, _ := fd.Options().(*descriptorpb.FieldOptions)
		if strings.HasSuffix(string(fd.Name()), "_deprecated") || opts.GetDeprecated() {
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
