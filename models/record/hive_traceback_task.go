// Code generated by github.com/actgardner/gogen-avro. DO NOT EDIT.
/*
 * SOURCES:
 *     HiveTask.avsc
 *     HiveTracebackTask.avsc
 *     OssTask.avsc
 *     OssTaskResult.avsc
 */

package record

import (
	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/container"
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/types"
	"io"
)

type HiveTracebackTask struct {
	JobId           string
	TraceId         string
	DatasetId       string
	ParentDatasetId []string
	ParentUrl       *ParentUrlRecord
	Length          int32
	TaskType        string
	Remarks         string
}

func NewHiveTracebackTaskWriter(writer io.Writer, codec container.Codec, recordsPerBlock int64) (*container.Writer, error) {
	str := &HiveTracebackTask{}
	return container.NewWriter(writer, codec, recordsPerBlock, str.Schema())
}

func DeserializeHiveTracebackTask(r io.Reader) (*HiveTracebackTask, error) {
	t := NewHiveTracebackTask()

	deser, err := compiler.CompileSchemaBytes([]byte(t.Schema()), []byte(t.Schema()))
	if err != nil {
		return nil, err
	}

	err = vm.Eval(r, deser, t)
	return t, err
}

func NewHiveTracebackTask() *HiveTracebackTask {
	return &HiveTracebackTask{}
}

func (r *HiveTracebackTask) Schema() string {
	return "{\"fields\":[{\"name\":\"jobId\",\"type\":\"string\"},{\"name\":\"traceId\",\"type\":\"string\"},{\"name\":\"datasetId\",\"type\":\"string\"},{\"name\":\"parentDatasetId\",\"type\":{\"items\":\"string\",\"type\":\"array\"}},{\"name\":\"parentUrl\",\"type\":{\"fields\":[{\"name\":\"MetaData\",\"type\":\"string\"},{\"name\":\"SampleData\",\"type\":\"string\"}],\"name\":\"ParentUrlRecord\",\"type\":\"record\"}},{\"name\":\"length\",\"type\":\"int\"},{\"name\":\"taskType\",\"type\":\"string\"},{\"name\":\"remarks\",\"type\":\"string\"}],\"name\":\"HiveTracebackTask\",\"namespace\":\"com.pharbers.kafka.schema\",\"type\":\"record\"}"
}

func (r *HiveTracebackTask) SchemaName() string {
	return "com.pharbers.kafka.schema.HiveTracebackTask"
}

func (r *HiveTracebackTask) Serialize(w io.Writer) error {
	return writeHiveTracebackTask(r, w)
}

func (_ *HiveTracebackTask) SetBoolean(v bool)    { panic("Unsupported operation") }
func (_ *HiveTracebackTask) SetInt(v int32)       { panic("Unsupported operation") }
func (_ *HiveTracebackTask) SetLong(v int64)      { panic("Unsupported operation") }
func (_ *HiveTracebackTask) SetFloat(v float32)   { panic("Unsupported operation") }
func (_ *HiveTracebackTask) SetDouble(v float64)  { panic("Unsupported operation") }
func (_ *HiveTracebackTask) SetBytes(v []byte)    { panic("Unsupported operation") }
func (_ *HiveTracebackTask) SetString(v string)   { panic("Unsupported operation") }
func (_ *HiveTracebackTask) SetUnionElem(v int64) { panic("Unsupported operation") }
func (r *HiveTracebackTask) Get(i int) types.Field {
	switch i {
	case 0:
		return (*types.String)(&r.JobId)
	case 1:
		return (*types.String)(&r.TraceId)
	case 2:
		return (*types.String)(&r.DatasetId)
	case 3:
		r.ParentDatasetId = make([]string, 0)
		return (*ArrayStringWrapper)(&r.ParentDatasetId)
	case 4:
		r.ParentUrl = NewParentUrlRecord()
		return r.ParentUrl
	case 5:
		return (*types.Int)(&r.Length)
	case 6:
		return (*types.String)(&r.TaskType)
	case 7:
		return (*types.String)(&r.Remarks)

	}
	panic("Unknown field index")
}
func (r *HiveTracebackTask) SetDefault(i int) {
	switch i {

	}
	panic("Unknown field index")
}
func (_ *HiveTracebackTask) AppendMap(key string) types.Field { panic("Unsupported operation") }
func (_ *HiveTracebackTask) AppendArray() types.Field         { panic("Unsupported operation") }
func (_ *HiveTracebackTask) Finalize()                        {}

type HiveTracebackTaskReader struct {
	r io.Reader
	p *vm.Program
}

func NewHiveTracebackTaskReader(r io.Reader) (*HiveTracebackTaskReader, error) {
	containerReader, err := container.NewReader(r)
	if err != nil {
		return nil, err
	}

	t := NewHiveTracebackTask()
	deser, err := compiler.CompileSchemaBytes([]byte(containerReader.AvroContainerSchema()), []byte(t.Schema()))
	if err != nil {
		return nil, err
	}

	return &HiveTracebackTaskReader{
		r: containerReader,
		p: deser,
	}, nil
}

func (r *HiveTracebackTaskReader) Read() (*HiveTracebackTask, error) {
	t := NewHiveTracebackTask()
	err := vm.Eval(r.r, r.p, t)
	return t, err
}
