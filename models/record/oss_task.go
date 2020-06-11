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

type OssTask struct {
	AssetId    string	`json:"AssetId"`
	JobId      string	`json:"JobId"`
	TraceId    string	`json:"TraceId"`
	OssKey     string	`json:"OssKey"`
	FileType   string	`json:"FileType"`
	FileName   string	`json:"FileName"`
	SheetName  string	`json:"SheetName"`
	Owner      string	`json:"Owner"`
	CreateTime int64	`json:"CreateTime"`
	Labels     []string	`json:"Labels"`
	DataCover  []string	`json:"DataCover"`
	GeoCover   []string	`json:"GeoCover"`
	Markets    []string	`json:"Markets"`
	Molecules  []string	`json:"Molecules"`
	Providers  []string	`json:"Providers"`
}

func NewOssTaskWriter(writer io.Writer, codec container.Codec, recordsPerBlock int64) (*container.Writer, error) {
	str := &OssTask{}
	return container.NewWriter(writer, codec, recordsPerBlock, str.Schema())
}

func DeserializeOssTask(r io.Reader) (*OssTask, error) {
	t := NewOssTask()

	deser, err := compiler.CompileSchemaBytes([]byte(t.Schema()), []byte(t.Schema()))
	if err != nil {
		return nil, err
	}

	err = vm.Eval(r, deser, t)
	return t, err
}

func NewOssTask() *OssTask {
	return &OssTask{}
}

func (r *OssTask) Schema() string {
	return "{\"fields\":[{\"name\":\"assetId\",\"type\":\"string\"},{\"name\":\"jobId\",\"type\":\"string\"},{\"name\":\"traceId\",\"type\":\"string\"},{\"name\":\"ossKey\",\"type\":\"string\"},{\"name\":\"fileType\",\"type\":\"string\"},{\"name\":\"fileName\",\"type\":\"string\"},{\"name\":\"sheetName\",\"type\":\"string\"},{\"name\":\"owner\",\"type\":\"string\"},{\"name\":\"createTime\",\"type\":\"long\"},{\"name\":\"labels\",\"type\":{\"items\":\"string\",\"type\":\"array\"}},{\"name\":\"dataCover\",\"type\":{\"items\":\"string\",\"type\":\"array\"}},{\"name\":\"geoCover\",\"type\":{\"items\":\"string\",\"type\":\"array\"}},{\"name\":\"markets\",\"type\":{\"items\":\"string\",\"type\":\"array\"}},{\"name\":\"molecules\",\"type\":{\"items\":\"string\",\"type\":\"array\"}},{\"name\":\"providers\",\"type\":{\"items\":\"string\",\"type\":\"array\"}}],\"name\":\"OssTask\",\"namespace\":\"com.pharbers.kafka.schema\",\"type\":\"record\"}"
}

func (r *OssTask) SchemaName() string {
	return "com.pharbers.kafka.schema.OssTask"
}

func (r *OssTask) Serialize(w io.Writer) error {
	return writeOssTask(r, w)
}

func (_ *OssTask) SetBoolean(v bool)    { panic("Unsupported operation") }
func (_ *OssTask) SetInt(v int32)       { panic("Unsupported operation") }
func (_ *OssTask) SetLong(v int64)      { panic("Unsupported operation") }
func (_ *OssTask) SetFloat(v float32)   { panic("Unsupported operation") }
func (_ *OssTask) SetDouble(v float64)  { panic("Unsupported operation") }
func (_ *OssTask) SetBytes(v []byte)    { panic("Unsupported operation") }
func (_ *OssTask) SetString(v string)   { panic("Unsupported operation") }
func (_ *OssTask) SetUnionElem(v int64) { panic("Unsupported operation") }
func (r *OssTask) Get(i int) types.Field {
	switch i {
	case 0:
		return (*types.String)(&r.AssetId)
	case 1:
		return (*types.String)(&r.JobId)
	case 2:
		return (*types.String)(&r.TraceId)
	case 3:
		return (*types.String)(&r.OssKey)
	case 4:
		return (*types.String)(&r.FileType)
	case 5:
		return (*types.String)(&r.FileName)
	case 6:
		return (*types.String)(&r.SheetName)
	case 7:
		return (*types.String)(&r.Owner)
	case 8:
		return (*types.Long)(&r.CreateTime)
	case 9:
		r.Labels = make([]string, 0)
		return (*ArrayStringWrapper)(&r.Labels)
	case 10:
		r.DataCover = make([]string, 0)
		return (*ArrayStringWrapper)(&r.DataCover)
	case 11:
		r.GeoCover = make([]string, 0)
		return (*ArrayStringWrapper)(&r.GeoCover)
	case 12:
		r.Markets = make([]string, 0)
		return (*ArrayStringWrapper)(&r.Markets)
	case 13:
		r.Molecules = make([]string, 0)
		return (*ArrayStringWrapper)(&r.Molecules)
	case 14:
		r.Providers = make([]string, 0)
		return (*ArrayStringWrapper)(&r.Providers)

	}
	panic("Unknown field index")
}
func (r *OssTask) SetDefault(i int) {
	switch i {

	}
	panic("Unknown field index")
}
func (_ *OssTask) AppendMap(key string) types.Field { panic("Unsupported operation") }
func (_ *OssTask) AppendArray() types.Field         { panic("Unsupported operation") }
func (_ *OssTask) Finalize()                        {}

type OssTaskReader struct {
	r io.Reader
	p *vm.Program
}

func NewOssTaskReader(r io.Reader) (*OssTaskReader, error) {
	containerReader, err := container.NewReader(r)
	if err != nil {
		return nil, err
	}

	t := NewOssTask()
	deser, err := compiler.CompileSchemaBytes([]byte(containerReader.AvroContainerSchema()), []byte(t.Schema()))
	if err != nil {
		return nil, err
	}

	return &OssTaskReader{
		r: containerReader,
		p: deser,
	}, nil
}

func (r *OssTaskReader) Read() (*OssTask, error) {
	t := NewOssTask()
	err := vm.Eval(r.r, r.p, t)
	return t, err
}
