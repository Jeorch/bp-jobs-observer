// Code generated by github.com/actgardner/gogen-avro. DO NOT EDIT.
/*
 * SOURCE:
 *     EventMsg.avsc
 */

package PhEventMsg

import (
	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/container"
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/types"
	"io"
)

type EventMsg struct {
	JobId   string
	TraceId string
	Type    string
	Data    string
}

func NewEventMsgWriter(writer io.Writer, codec container.Codec, recordsPerBlock int64) (*container.Writer, error) {
	str := &EventMsg{}
	return container.NewWriter(writer, codec, recordsPerBlock, str.Schema())
}

func DeserializeEventMsg(r io.Reader) (*EventMsg, error) {
	t := NewEventMsg()
	err := deserializeField(r, t.Schema(), t.Schema(), t)
	return t, err
}

func DeserializeEventMsgFromSchema(r io.Reader, schema string) (*EventMsg, error) {
	t := NewEventMsg()
	err := deserializeField(r, schema, t.Schema(), t)
	return t, err
}

func NewEventMsg() *EventMsg {
	return &EventMsg{}
}

func (r *EventMsg) Schema() string {
	return "{\"fields\":[{\"name\":\"jobId\",\"type\":\"string\"},{\"name\":\"traceId\",\"type\":\"string\"},{\"name\":\"type\",\"type\":\"string\"},{\"name\":\"data\",\"type\":\"string\"}],\"name\":\"EventMsg\",\"namespace\":\"com.pharbers.kafka.schema\",\"type\":\"record\"}"
}

func (r *EventMsg) SchemaName() string {
	return "com.pharbers.kafka.schema.EventMsg"
}

func (r *EventMsg) Serialize(w io.Writer) error {
	return writeEventMsg(r, w)
}

func (_ *EventMsg) SetBoolean(v bool)    { panic("Unsupported operation") }
func (_ *EventMsg) SetInt(v int32)       { panic("Unsupported operation") }
func (_ *EventMsg) SetLong(v int64)      { panic("Unsupported operation") }
func (_ *EventMsg) SetFloat(v float32)   { panic("Unsupported operation") }
func (_ *EventMsg) SetDouble(v float64)  { panic("Unsupported operation") }
func (_ *EventMsg) SetBytes(v []byte)    { panic("Unsupported operation") }
func (_ *EventMsg) SetString(v string)   { panic("Unsupported operation") }
func (_ *EventMsg) SetUnionElem(v int64) { panic("Unsupported operation") }
func (r *EventMsg) Get(i int) types.Field {
	switch i {
	case 0:
		return (*types.String)(&r.JobId)
	case 1:
		return (*types.String)(&r.TraceId)
	case 2:
		return (*types.String)(&r.Type)
	case 3:
		return (*types.String)(&r.Data)

	}
	panic("Unknown field index")
}
func (r *EventMsg) SetDefault(i int) {
	switch i {

	}
	panic("Unknown field index")
}
func (_ *EventMsg) AppendMap(key string) types.Field { panic("Unsupported operation") }
func (_ *EventMsg) AppendArray() types.Field         { panic("Unsupported operation") }
func (_ *EventMsg) Finalize()                        {}

type EventMsgReader struct {
	r io.Reader
	p *vm.Program
}

func NewEventMsgReader(r io.Reader) (*EventMsgReader, error) {
	containerReader, err := container.NewReader(r)
	if err != nil {
		return nil, err
	}

	t := NewEventMsg()
	deser, err := compiler.CompileSchemaBytes([]byte(containerReader.AvroContainerSchema()), []byte(t.Schema()))
	if err != nil {
		return nil, err
	}

	return &EventMsgReader{
		r: containerReader,
		p: deser,
	}, nil
}

func (r *EventMsgReader) Read() (*EventMsg, error) {
	t := NewEventMsg()
	err := vm.Eval(r.r, r.p, t)
	return t, err
}