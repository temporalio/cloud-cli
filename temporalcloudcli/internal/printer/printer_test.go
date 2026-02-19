package printer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"go.temporal.io/cloud-sdk/api/resource/v1"
)

func TestPrinter_Text(t *testing.T) {
	type MyStruct struct {
		Foo              string
		Bar              bool
		unexportedBaz    string
		ReallyLongField  any
		Omitted          string `cli:",omit"`
		OmittedCardEmpty string `cli:",cardOmitEmpty"`
	}
	var buf bytes.Buffer
	p := Printer{Output: &buf}
	// Simple struct non-table no fields set
	require.NoError(t, p.PrintStructured([]*MyStruct{
		{
			Foo:           "1",
			unexportedBaz: "2",
			ReallyLongField: struct {
				Key any `json:"key"`
			}{Key: 123},
			Omitted:          "value",
			OmittedCardEmpty: "value",
		},
		{
			Foo:             "not-a-number",
			Bar:             true,
			ReallyLongField: map[string]int{"": 0},
		},
	}, StructuredOptions{}))
	// Check
	require.Equal(t, normalizeMultiline(`
  Foo               1
  Bar               false
  ReallyLongField   {"key":123}
  OmittedCardEmpty  value

  Foo              not-a-number
  Bar              true
  ReallyLongField  map[:0]`), normalizeMultiline(buf.String()))
}

func normalizeMultiline(s string) string {
	// Split lines, trim trailing space on each (also removes \r), remove empty
	// lines, re-join
	var ret string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		// Only non-empty lines
		if line != "" {
			if ret != "" {
				ret += "\n"
			}
			ret += line
		}
	}
	return ret
}

func TestPrinter_JSON(t *testing.T) {
	var buf bytes.Buffer

	// With indentation
	p := Printer{Output: &buf, JSON: true, JSONIndent: "  "}
	p.Println("should not print")
	require.NoError(t, p.PrintStructured(map[string]string{"foo": "bar"}, StructuredOptions{}))
	require.Equal(t, `{
  "foo": "bar"
}
`, buf.String())

	// Without indentation
	buf.Reset()
	p = Printer{Output: &buf, JSON: true}
	p.Println("should not print")
	require.NoError(t, p.PrintStructured(map[string]string{"foo": "bar"}, StructuredOptions{}))
	require.Equal(t, "{\"foo\":\"bar\"}\n", buf.String())
}

func TestPrinter_Table(t *testing.T) {
	type TableStruct struct {
		Name   string
		Age    int
		Active bool
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	// Test table with header
	require.NoError(t, p.PrintStructured([]TableStruct{
		{Name: "Alice", Age: 30, Active: true},
		{Name: "Bob", Age: 25, Active: false},
	}, StructuredOptions{
		Table: &TableOptions{},
	}))
	require.Contains(t, buf.String(), "Name")
	require.Contains(t, buf.String(), "Age")
	require.Contains(t, buf.String(), "Alice")
	require.Contains(t, buf.String(), "Bob")

	// Test table without header
	buf.Reset()
	require.NoError(t, p.PrintStructured([]TableStruct{
		{Name: "Charlie", Age: 35, Active: true},
	}, StructuredOptions{
		Table: &TableOptions{NoHeader: true},
	}))
	require.NotContains(t, buf.String(), "Name")
	require.Contains(t, buf.String(), "Charlie")
}

func TestPrinter_TableWithOptions(t *testing.T) {
	type TableStruct struct {
		Short string
		Long  string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	// Test with field widths and alignment
	require.NoError(t, p.PrintStructured([]TableStruct{
		{Short: "A", Long: "This is a very long string"},
	}, StructuredOptions{
		Table: &TableOptions{
			FieldWidths: map[string]int{"Short": 10, "Long": 15},
			FieldAlign:  map[string]Align{"Short": AlignCenter},
		},
	}))
	output := buf.String()
	require.Contains(t, output, "A")
	require.Contains(t, output, "This is a very")
}

func TestPrinter_SpecificFields(t *testing.T) {
	type MyStruct struct {
		Field1 string
		Field2 string
		Field3 string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	// Test with specific fields
	require.NoError(t, p.PrintStructured(MyStruct{
		Field1: "value1",
		Field2: "value2",
		Field3: "value3",
	}, StructuredOptions{
		Fields: []string{"Field1", "Field3"},
	}))
	require.Contains(t, buf.String(), "Field1")
	require.Contains(t, buf.String(), "Field3")
	require.NotContains(t, buf.String(), "Field2")
}

func TestPrinter_ExcludeFields(t *testing.T) {
	type MyStruct struct {
		Include string
		Exclude string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	require.NoError(t, p.PrintStructured(MyStruct{
		Include: "show",
		Exclude: "hide",
	}, StructuredOptions{
		ExcludeFields: []string{"Exclude"},
	}))
	require.Contains(t, buf.String(), "Include")
	require.NotContains(t, buf.String(), "Exclude")
}

func TestPrinter_TextVal(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{Output: &buf}

	type TestStruct struct {
		TimeField   time.Time
		ByteField   []byte
		SliceField  []string
		IntField    int
		FloatField  float64
		StructField struct{ X int }
	}

	now := time.Now()
	testData := TestStruct{
		TimeField:   now,
		ByteField:   []byte{1, 2, 3},
		SliceField:  []string{"a", "b", "c"},
		IntField:    42,
		FloatField:  3.14,
		StructField: struct{ X int }{X: 10},
	}

	require.NoError(t, p.PrintStructured(testData, StructuredOptions{}))
	output := buf.String()

	// Check that time is formatted
	require.Contains(t, output, now.Format(time.RFC3339))
	// Check that bytes are base64 encoded
	require.Contains(t, output, "bytes(")
	// Check that slice is formatted with brackets
	require.Contains(t, output, "[a, b, c]")
	// Check numbers
	require.Contains(t, output, "42")
	require.Contains(t, output, "3.14")
}

func TestPrinter_CustomTimeFormat(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{
		Output: &buf,
		FormatTime: func(t time.Time) string {
			return "custom-" + t.Format("2006")
		},
	}

	type TimeStruct struct {
		Created time.Time
	}

	testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, p.PrintStructured(TimeStruct{Created: testTime}, StructuredOptions{}))
	require.Contains(t, buf.String(), "custom-2023")
}

func TestPrinter_JSONList(t *testing.T) {
	var buf bytes.Buffer

	// With indentation
	p := Printer{Output: &buf, JSON: true, JSONIndent: "  "}
	p.StartList()
	p.Println("should not print")
	require.NoError(t, p.PrintStructured(map[string]string{"foo": "bar"}, StructuredOptions{}))
	require.NoError(t, p.PrintStructured(map[string]string{"baz": "qux"}, StructuredOptions{}))
	p.EndList()
	require.Equal(t, `[
{
  "foo": "bar"
},
{
  "baz": "qux"
}
]
`, buf.String())

	// Without indentation
	buf.Reset()
	p = Printer{Output: &buf, JSON: true}
	p.StartList()
	p.Println("should not print")
	require.NoError(t, p.PrintStructured(map[string]string{"foo": "bar"}, StructuredOptions{}))
	require.NoError(t, p.PrintStructured(map[string]string{"baz": "qux"}, StructuredOptions{}))
	p.EndList()
	require.Equal(t, "{\"foo\":\"bar\"}\n{\"baz\":\"qux\"}\n", buf.String())

	// Empty with indentation
	buf.Reset()
	p = Printer{Output: &buf, JSON: true, JSONIndent: "  "}
	p.StartList()
	p.Println("should not print")
	p.EndList()
	require.Equal(t, "[\n]\n", buf.String())

	// Empty without indentation
	buf.Reset()
	p = Printer{Output: &buf, JSON: true}
	p.StartList()
	p.Println("should not print")
	p.EndList()
	require.Equal(t, "", buf.String())
}

func TestPrinter_PrintAndPrintln(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{Output: &buf}

	p.Print("hello")
	p.Print(" ", "world")
	p.Println()
	p.Printlnf("test %d", 123)

	require.Equal(t, "hello world\ntest 123\n", buf.String())

	// Verify Print/Println are ignored in JSON mode
	buf.Reset()
	p.JSON = true
	p.Print("should not appear")
	p.Println("also should not appear")
	require.Empty(t, buf.String())
}

func TestPrinter_PrintDiff(t *testing.T) {
	type DiffStruct struct {
		Name  string
		Value int
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	a := DiffStruct{Name: "old", Value: 10}
	b := DiffStruct{Name: "new", Value: 20}

	require.NoError(t, p.PrintDiff(a, b, DiffOptions{Verbose: true}))
	output := buf.String()
	require.Contains(t, output, "old")
	require.Contains(t, output, "new")

	// Test with NoColor option
	buf.Reset()
	require.NoError(t, p.PrintDiff(a, b, DiffOptions{NoColor: true, Verbose: true}))
	require.NotEmpty(t, buf.String())

	// Test non-verbose mode
	buf.Reset()
	require.NoError(t, p.PrintDiff(a, b, DiffOptions{Verbose: false}))
	require.NotEmpty(t, buf.String())
}

func TestPrinter_PrintDiff_JSON(t *testing.T) {
	type DiffStruct struct {
		Name  string
		Value int
	}

	a := DiffStruct{Name: "old", Value: 10}
	b := DiffStruct{Name: "new", Value: 20}

	// JSONL mode
	var buf bytes.Buffer
	p := Printer{Output: &buf, JSON: true}
	require.NoError(t, p.PrintDiff(a, b, DiffOptions{}))
	var result map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Contains(t, result, "before")
	require.Contains(t, result, "after")
	require.Contains(t, string(result["before"]), "old")
	require.Contains(t, string(result["after"]), "new")

	// Pretty JSON mode
	buf.Reset()
	p.JSONIndent = "  "
	require.NoError(t, p.PrintDiff(a, b, DiffOptions{}))
	var prettyResult map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(buf.Bytes(), &prettyResult))
	require.Contains(t, prettyResult, "before")
	require.Contains(t, prettyResult, "after")
}

func TestPrinter_PrintDiff_DifferentTypes(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{Output: &buf}

	err := p.PrintDiff("string", 123, DiffOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot diff different types")
}

func TestPrinter_PrintResource(t *testing.T) {
	type ResourceSpec struct {
		Replicas int
		Image    string
	}

	type Resource struct {
		Name      string
		Namespace string
		Spec      ResourceSpec
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	resource := Resource{
		Name:      "my-resource",
		Namespace: "default",
		Spec: ResourceSpec{
			Replicas: 3,
			Image:    "nginx:latest",
		},
	}

	// Test text mode
	require.NoError(t, p.PrintResource(resource, PrintResourceOptions{}))
	output := buf.String()
	require.Contains(t, output, "Name")
	require.Contains(t, output, "my-resource")
	require.Contains(t, output, "Namespace")
	require.Contains(t, output, "default")
	require.Contains(t, output, "Spec:")
	require.Contains(t, output, "Replicas")
	require.Contains(t, output, "3")
	require.Contains(t, output, "Image")
	require.Contains(t, output, "nginx:latest")

	// Test with specific fields
	buf.Reset()
	require.NoError(t, p.PrintResource(resource, PrintResourceOptions{
		Fields:     []string{"Name"},
		SpecFields: []string{"Replicas"},
	}))
	output = buf.String()
	require.Contains(t, output, "Name")
	require.Contains(t, output, "my-resource")
	require.NotContains(t, output, "Namespace")
	require.Contains(t, output, "Replicas")
	require.NotContains(t, output, "Image")

	// Test JSON mode
	buf.Reset()
	p.JSON = true
	p.JSONIndent = "  "
	require.NoError(t, p.PrintResource(resource, PrintResourceOptions{}))
	var jsonResult map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &jsonResult))
	require.Equal(t, "my-resource", jsonResult["Name"])
}

func TestPrinter_PrintResource_Pointer(t *testing.T) {
	type ResourceSpec struct {
		Count int
	}

	type Resource struct {
		ID   string
		Spec *ResourceSpec
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	resource := Resource{
		ID:   "test-id",
		Spec: &ResourceSpec{Count: 5},
	}

	require.NoError(t, p.PrintResource(&resource, PrintResourceOptions{}))
	require.Contains(t, buf.String(), "test-id")
	require.Contains(t, buf.String(), "5")
}

func TestPrinter_PrintResource_NilSpec(t *testing.T) {
	type ResourceSpec struct {
		Value string
	}

	type Resource struct {
		Name string
		Spec *ResourceSpec
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	resource := Resource{
		Name: "test",
		Spec: nil,
	}

	require.NoError(t, p.PrintResource(resource, PrintResourceOptions{}))
	require.Contains(t, buf.String(), "test")
}

func TestPrinter_PrintResourceList(t *testing.T) {
	type ResourceSpec struct {
		Size int
	}

	type Resource struct {
		Name string
		Spec ResourceSpec
	}

	type ResourceListResponse struct {
		Resources     []Resource
		NextPageToken string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	response := ResourceListResponse{
		Resources: []Resource{
			{Name: "resource1", Spec: ResourceSpec{Size: 10}},
			{Name: "resource2", Spec: ResourceSpec{Size: 20}},
		},
		NextPageToken: "next-token-123",
	}

	// Test text mode with table
	require.NoError(t, p.PrintResourceList(response, PrintResourceOptions{}, TableOptions{}))
	output := buf.String()
	require.Contains(t, output, "resource1")
	require.Contains(t, output, "resource2")
	require.Contains(t, output, "Next page token: next-token-123")

	// Test JSON mode
	buf.Reset()
	p.JSON = true
	require.NoError(t, p.PrintResourceList(response, PrintResourceOptions{}, TableOptions{}))
	var jsonResult map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &jsonResult))
	require.NotNil(t, jsonResult["Resources"])
}

func TestPrinter_PrintResourceList_EmptyToken(t *testing.T) {
	type Resource struct {
		ID string
	}

	type ResourceListResponse struct {
		Resources     []Resource
		NextPageToken string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	response := ResourceListResponse{
		Resources:     []Resource{{ID: "test"}},
		NextPageToken: "",
	}

	require.NoError(t, p.PrintResourceList(response, PrintResourceOptions{}, TableOptions{}))
	require.NotContains(t, buf.String(), "Next page token")
}

func TestPrinter_PrintResourceList_Pointer(t *testing.T) {
	type Resource struct {
		Name string
	}

	type ResourceListResponse struct {
		Resources     []*Resource
		NextPageToken string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	response := &ResourceListResponse{
		Resources: []*Resource{
			{Name: "ptr-resource"},
		},
	}

	require.NoError(t, p.PrintResourceList(response, PrintResourceOptions{}, TableOptions{}))
	require.Contains(t, buf.String(), "ptr-resource")
}

func TestPrinter_PrintResourceList_FirstResourceEmptySpec(t *testing.T) {
	type ResourceSpec struct {
		Region string
	}

	type Resource struct {
		Name string
		Spec ResourceSpec
	}

	type ResourceListResponse struct {
		Resources     []Resource
		NextPageToken string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	// First resource has an empty Spec; second has a populated one.
	// The Region column must still appear for both rows.
	response := ResourceListResponse{
		Resources: []Resource{
			{Name: "resource1"},
			{Name: "resource2", Spec: ResourceSpec{Region: "us-east-1"}},
		},
	}

	require.NoError(t, p.PrintResourceList(response, PrintResourceOptions{}, TableOptions{}))
	output := buf.String()
	require.Contains(t, output, "Region", "spec column header must be present even when first resource has empty spec")
	require.Contains(t, output, "us-east-1")
}

func TestPrinter_StructTags(t *testing.T) {
	type TaggedStruct struct {
		Normal        string
		Width10       string `cli:",width=10"`
		AlignRight    string `cli:",align=right"`
		AlignCenter   string `cli:",align=center"`
		AlignLeft     string `cli:",align=left"`
		CardOmitEmpty string `cli:",cardOmitEmpty"`
		JSONOmitEmpty string `json:",omitempty"`
		BothOmitEmpty string `cli:",cardOmitEmpty" json:",omitempty"`
		OmittedField  string `cli:",omit"`
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	data := TaggedStruct{
		Normal:      "normal",
		Width10:     "w10",
		AlignRight:  "right",
		AlignCenter: "center",
		AlignLeft:   "left",
		// Leave CardOmitEmpty, JSONOmitEmpty, BothOmitEmpty empty to test omission
		OmittedField: "should-not-appear",
	}

	require.NoError(t, p.PrintStructured(data, StructuredOptions{}))
	output := buf.String()
	require.Contains(t, output, "normal")
	require.Contains(t, output, "right")
	require.NotContains(t, output, "should-not-appear")
	require.NotContains(t, output, "CardOmitEmpty")
	require.NotContains(t, output, "JSONOmitEmpty")
	require.NotContains(t, output, "BothOmitEmpty")
}

func TestPrinter_NumericTypes(t *testing.T) {
	type NumericStruct struct {
		Int     int
		Int8    int8
		Int16   int16
		Int32   int32
		Int64   int64
		Uint    uint
		Uint8   uint8
		Uint16  uint16
		Uint32  uint32
		Uint64  uint64
		Float32 float32
		Float64 float64
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	data := NumericStruct{
		Int: 1, Int8: 2, Int16: 3, Int32: 4, Int64: 5,
		Uint: 6, Uint8: 7, Uint16: 8, Uint32: 9, Uint64: 10,
		Float32: 1.5, Float64: 2.5,
	}

	require.NoError(t, p.PrintStructured(data, StructuredOptions{
		Table: &TableOptions{},
	}))
	output := buf.String()
	// Verify field names appear
	require.Contains(t, output, "Int")
	require.Contains(t, output, "Float32")
	require.Contains(t, output, "Float64")
	// Verify some numeric values appear
	require.Contains(t, output, "1")
	require.Contains(t, output, "2.5")
	require.Contains(t, output, "1.5")
}

func TestPrinter_NilElementsInPointerSlice(t *testing.T) {
	type Item struct {
		Name  string
		Value int
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	// A nil element mixed with non-nil elements must not panic.
	items := []*Item{
		{Name: "first", Value: 1},
		nil,
		{Name: "third", Value: 3},
	}

	// Table mode
	require.NoError(t, p.PrintStructured(items, StructuredOptions{Table: &TableOptions{}}))
	output := buf.String()
	require.Contains(t, output, "first")
	require.Contains(t, output, "third")

	// Card mode
	buf.Reset()
	require.NoError(t, p.PrintStructured(items, StructuredOptions{}))
	output = buf.String()
	require.Contains(t, output, "first")
	require.Contains(t, output, "third")
}

func TestPrinter_MapData(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{Output: &buf}

	data := []map[string]any{
		{"Name": "Alice", "Age": 30},
		{"Name": "Bob", "Age": 25},
	}

	require.NoError(t, p.PrintStructured(data, StructuredOptions{
		Fields: []string{"Name", "Age"},
		Table:  &TableOptions{},
	}))
	output := buf.String()
	require.Contains(t, output, "Alice")
	require.Contains(t, output, "Bob")
}

func TestPrinter_SingleItem(t *testing.T) {
	type Item struct {
		Value string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	// Single item (not in a slice)
	require.NoError(t, p.PrintStructured(Item{Value: "single"}, StructuredOptions{}))
	require.Contains(t, buf.String(), "single")
}

func TestPrinter_ExtraIndent(t *testing.T) {
	type Simple struct {
		Field string
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	require.NoError(t, p.PrintStructured(Simple{Field: "value"}, StructuredOptions{
		NonJSONExtraIndent: 2,
	}))
	output := buf.String()
	// Should have extra indentation (3 levels total: base + extra 2)
	require.Contains(t, output, "      ") // 6 spaces = 3 indents * 2 spaces
}

func TestPrinter_ZeroTimeHandling(t *testing.T) {
	type TimeStruct struct {
		Created time.Time
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	// Zero time should render as empty string
	require.NoError(t, p.PrintStructured(TimeStruct{}, StructuredOptions{}))
	output := buf.String()
	require.Contains(t, output, "Created")
	// The value part should be empty or whitespace
	lines := strings.Split(output, "\n")
	var createdLine string
	for _, line := range lines {
		if strings.Contains(line, "Created") {
			createdLine = line
			break
		}
	}
	require.NotEmpty(t, createdLine)
	// After "Created", there should be no timestamp value
	parts := strings.Fields(createdLine)
	require.Len(t, parts, 1) // Just "Created", no value
}

func TestPrinter_JSONMarshaler(t *testing.T) {
	type CustomJSON struct {
		Value int
	}

	// This type implements json.Marshaler
	data := struct {
		Custom CustomJSON
		Normal string
	}{
		Custom: CustomJSON{Value: 42},
		Normal: "text",
	}

	var buf bytes.Buffer
	p := Printer{Output: &buf}

	require.NoError(t, p.PrintStructured(data, StructuredOptions{}))
	output := buf.String()
	require.Contains(t, output, "42")
	require.Contains(t, output, "text")
}

func TestPrinter_OverrideJSONPayloadShorthand(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{Output: &buf, JSON: true, JSONPayloadShorthand: true}

	trueVal := true
	falseVal := false

	// Test with override to false
	require.NoError(t, p.PrintStructured(
		map[string]string{"key": "value"},
		StructuredOptions{OverrideJSONPayloadShorthand: &falseVal},
	))
	require.Contains(t, buf.String(), "key")

	// Test with override to true
	buf.Reset()
	p.JSONPayloadShorthand = false
	require.NoError(t, p.PrintStructured(
		map[string]string{"key": "value"},
		StructuredOptions{OverrideJSONPayloadShorthand: &trueVal},
	))
	require.Contains(t, buf.String(), "key")
}

func TestConverters_ResourceState(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{Output: &buf, JSON: true, JSONPayloadShorthand: true}
	RegisterEnumToStringConverter[resource.ResourceState](&p, "RESOURCE_STATE_", resource.ResourceState_name)

	tests := []struct {
		name   string
		input  any
		output string
		ok     bool
	}{
		{
			name:   "ACTIVE state",
			input:  resource.ResourceState_RESOURCE_STATE_ACTIVE,
			output: "ACTIVE",
			ok:     true,
		},
		{
			name:   "DELETED state",
			input:  resource.ResourceState_RESOURCE_STATE_DELETED,
			output: "DELETED",
			ok:     true,
		},
		{
			name:   "Invalid state returns UNKNOWN",
			input:  resource.ResourceState(999),
			output: "UNKNOWN",
			ok:     true,
		},
		{
			name:   "Wrong type returns not ok",
			input:  "not-a-resource-state",
			output: "",
			ok:     false,
		},
		{
			name:   "Integer type returns not ok",
			input:  42,
			output: "",
			ok:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := p.applyConverters(tt.input)
			assert.Equal(t, tt.ok, ok, "Converter ok status mismatch")
			if tt.ok {
				assert.Equal(t, tt.output, result, "Converter output mismatch")
			}
		})
	}
}

func TestConverters_AsyncOperationState(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{Output: &buf, JSON: true, JSONPayloadShorthand: true}
	RegisterEnumToStringConverter[operation.AsyncOperation_State](&p, "STATE_", operation.AsyncOperation_State_name)
	tests := []struct {
		name   string
		input  any
		output string
		ok     bool
	}{
		{
			name:   "PENDING state",
			input:  operation.AsyncOperation_STATE_PENDING,
			output: "PENDING",
			ok:     true,
		},
		{
			name:   "Invalid async operation state",
			input:  operation.AsyncOperation_State(888),
			output: "UNKNOWN",
			ok:     true,
		},
		{
			name:   "Wrong type returns not ok",
			input:  "not-an-operation-state",
			output: "",
			ok:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := p.applyConverters(tt.input)
			assert.Equal(t, tt.ok, ok, "Converter ok status mismatch")
			if tt.ok {
				assert.Equal(t, tt.output, result, "Converter output mismatch")
			}
		})
	}
}

func TestConverters_MultipleConverters(t *testing.T) {
	var buf bytes.Buffer
	p := Printer{Output: &buf, JSON: true, JSONPayloadShorthand: true}
	RegisterEnumToStringConverter[resource.ResourceState](&p, "RESOURCE_STATE_", resource.ResourceState_name)
	RegisterEnumToStringConverter[operation.AsyncOperation_State](&p, "STATE_", operation.AsyncOperation_State_name)

	// Test that both converters are registered and work independently
	resourceState := resource.ResourceState_RESOURCE_STATE_ACTIVE
	asyncState := operation.AsyncOperation_STATE_PENDING

	result1, ok1 := p.applyConverters(resourceState)
	require.True(t, ok1)
	require.Equal(t, "ACTIVE", result1)

	result2, ok2 := p.applyConverters(asyncState)
	require.True(t, ok2)
	require.Equal(t, "PENDING", result2)
}
