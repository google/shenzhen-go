package jspb

import "github.com/gopherjs/gopherjs/js"

// Writer encapsulates the jspb.BinaryWriter.
type Writer interface {
	GetResult() []byte

	// Scalars
	WriteInt32(int, int32)
	WriteInt64(int, int64)
	WriteUint32(int, uint32)
	WriteUint64(int, uint64)
	WriteSint32(int, int32)
	WriteSint64(int, int64)
	WriteFixed32(int, uint32)
	WriteFixed64(int, uint64)
	WriteSfixed32(int, int32)
	WriteSfixed64(int, int64)
	WriteFloat32(int, float32)
	WriteFloat64(int, float64)
	WriteEnum(int, int)
	WriteBool(int, bool)
	WriteString(int, string)
	WriteBytes(int, []byte)

	// Scalar Slices
	WriteInt32Slice(int, []int32)
	WriteInt64Slice(int, []int64)
	WriteUint32Slice(int, []uint32)
	WriteUint64Slice(int, []uint64)
	WriteSint32Slice(int, []int32)
	WriteSint64Slice(int, []int64)
	WriteFixed32Slice(int, []uint32)
	WriteFixed64Slice(int, []uint64)
	WriteSfixed32Slice(int, []int32)
	WriteSfixed64Slice(int, []int64)
	WriteFloat32Slice(int, []float32)
	WriteFloat64Slice(int, []float64)
	WriteBoolSlice(int, []bool)
	WriteEnumSlice(int, []int)

	// Specials
	WriteMessage(int, func())
}

type encoder struct {
	*js.Object
}

// WriteInt64 writes a signed 64bit varint
func (e encoder) WriteInt64(value int64) {
	e.Call("writeSplitVarint64", value&0xffffffff, uint64(value)>>32)
}

// WriteUint64 writes an unsigned 64bit varint
func (e encoder) WriteUint64(value uint64) {
	e.Call("writeSplitVarint64", value&0xffffffff, value>>32)
}

// WriteZigzag64 writes a Zigzag encoded signed 64bit varint
func (e encoder) WriteZigzag64(value int64) {
	// https://github.com/gogo/protobuf/blob/1ef32a8b9fc3f8ec940126907cedb5998f6318e4/proto/encode.go#L177
	v := (uint64(value) << 1) ^ uint64(value>>63)
	e.WriteUint64(v)
}

// WriteFixed64 writes an unsigned 64bit integer
func (e encoder) WriteFixed64(value uint64) {
	e.Call("writeSplitFixed64", value&0xffffffff, value>>32)
}

// WriteSignedFixed64 writes a signed 64bit varint
func (e encoder) WriteSignedFixed64(value int64) {
	e.Call("writeSplitFixed64", value&0xffffffff, uint64(value)>>32)
}

// NewWriter returns a new Writer ready for writing.
func NewWriter() Writer {
	w := &writer{
		Object: js.Global.Get("BinaryWriter").New(),
	}
	w.encoder = &encoder{
		Object: w.Get("encoder_"),
	}
	return w
}

// writer implements the Writer interface.
type writer struct {
	*js.Object
	encoder *encoder
}

type wireType int

// https://github.com/google/protobuf/blob/25625b956a2f0d432582009c16553a9fd21c3cea/js/binary/constants.js#L219
const (
	varint wireType = iota
	fixed64
	delimited
	startGroup
	endGroup
	fixed32
)

// GetResult returns the contents of the buffer as a byte slice.
func (w writer) GetResult() []byte {
	return w.Call("getResultBuffer").Interface().([]byte)
}

// WriteInt32 writes an int32 field to the buffer.
func (w writer) WriteInt32(field int, value int32) {
	w.Call("writeInt32", field, value)
}

// WriteInt64 writes an int64 field to the buffer.
func (w writer) WriteInt64(field int, value int64) {
	w.Call("writeFieldHeader_", field, varint)
	w.encoder.WriteInt64(value)
}

// WriteUint32 writes a uint32 field to the buffer.
func (w writer) WriteUint32(field int, value uint32) {
	w.Call("writeUint32", field, value)
}

// WriteUint64 writes a uint64 field to the buffer.
func (w writer) WriteUint64(field int, value uint64) {
	w.Call("writeFieldHeader_", field, varint)
	w.encoder.WriteUint64(value)
}

// WriteSint32 writes an sint32 field to the buffer.
func (w writer) WriteSint32(field int, value int32) {
	w.Call("writeSint32", field, value)
}

// WriteSint64 writes an sint64 field to the buffer.
func (w writer) WriteSint64(field int, value int64) {
	w.Call("writeFieldHeader_", field, varint)
	w.encoder.WriteZigzag64(value)
}

// WriteFixed32 writes a fixed32 field to the buffer.
func (w writer) WriteFixed32(field int, value uint32) {
	w.Call("writeFixed32", field, value)
}

// WriteFixed64 writes a fixed64 field to the buffer.
func (w writer) WriteFixed64(field int, value uint64) {
	w.Call("writeFieldHeader_", field, fixed64)
	w.encoder.WriteFixed64(value)
}

// WriteSfixed32 writes an sfixed32 field to the buffer.
func (w writer) WriteSfixed32(field int, value int32) {
	w.Call("writeSfixed32", field, value)
}

// WriteSfixed64 writes an sfixed64 field to the buffer.
func (w writer) WriteSfixed64(field int, value int64) {
	w.Call("writeFieldHeader_", field, fixed64)
	w.encoder.WriteSignedFixed64(value)
}

// WriteFloat32 writes a float32 field to the buffer
func (w writer) WriteFloat32(field int, value float32) {
	w.Call("writeFloat", field, value)
}

// WriteFloat64 writes a float64 field to the buffer
func (w writer) WriteFloat64(field int, value float64) {
	w.Call("writeDouble", field, value)
}

// WriteEnum writes an enum (as an int) to the buffer
func (w writer) WriteEnum(field int, value int) {
	w.Call("writeEnum", field, value)
}

// WriteBool writes a bool field to the buffer
func (w writer) WriteBool(field int, value bool) {
	w.Call("writeBool", field, value)
}

// WriteString writes a string field to the buffer
func (w writer) WriteString(field int, value string) {
	w.Call("writeString", field, value)
}

// WriteBytes writes a bytes field to the buffer
func (w writer) WriteBytes(field int, value []byte) {
	w.Call("writeBytes", field, value)
}

// WriteMessage writes a message to the buffer using writeFunc
func (w writer) WriteMessage(field int, writeFunc func()) {
	w.Call("writeMessage", field, 0 /* Unused */, writeFunc)
}

// WriteInt32Slice writes a repeated int32 field to the buffer.
func (w writer) WriteInt32Slice(field int, values []int32) {
	w.Call("writePackedInt32", field, values)
}

// WriteInt64Slice writes a repeated int64 field to the buffer.
func (w writer) WriteInt64Slice(field int, values []int64) {
	b := w.Call("beginDelimited_", field)
	for _, value := range values {
		w.encoder.WriteInt64(value)
	}
	w.Call("endDelimited_", b)
}

// WriteUint32Slice writes a repeated uint32 field to the buffer.
func (w writer) WriteUint32Slice(field int, values []uint32) {
	w.Call("writePackedUint32", field, values)
}

// WriteUint64Slice writes a repeated uint64 field to the buffer.
func (w writer) WriteUint64Slice(field int, values []uint64) {
	b := w.Call("beginDelimited_", field)
	for _, value := range values {
		w.encoder.WriteUint64(value)
	}
	w.Call("endDelimited_", b)
}

// WriteSint32Slice writes a repeated sint32 field to the buffer.
func (w writer) WriteSint32Slice(field int, values []int32) {
	w.Call("writePackedSint32", field, values)
}

// WriteSint64Slice writes a repeated sint64 field to the buffer.
func (w writer) WriteSint64Slice(field int, values []int64) {
	b := w.Call("beginDelimited_", field)
	for _, value := range values {
		w.encoder.WriteZigzag64(value)
	}
	w.Call("endDelimited_", b)
}

// WriteFixed32Slice writes a repeated fixed32 field to the buffer.
func (w writer) WriteFixed32Slice(field int, values []uint32) {
	w.Call("writePackedFixed32", field, values)
}

// WriteFixed64Slice writes a repeated fixed64 field to the buffer.
func (w writer) WriteFixed64Slice(field int, values []uint64) {
	w.Call("writeFieldHeader_", field, delimited)
	e := w.Get("encoder_")
	e.Call("writeUnsignedVarint32", len(values)*8)
	for _, value := range values {
		w.encoder.WriteFixed64(value)
	}
}

// WriteSfixed32Slice writes a repeated sfixed32 field to the buffer.
func (w writer) WriteSfixed32Slice(field int, values []int32) {
	w.Call("writePackedSfixed32", field, values)
}

// WriteSfixed64Slice writes a repeated sfixed64 field to the buffer.
func (w writer) WriteSfixed64Slice(field int, values []int64) {
	w.Call("writeFieldHeader_", field, delimited)
	e := w.Get("encoder_")
	e.Call("writeUnsignedVarint32", len(values)*8)
	for _, value := range values {
		w.encoder.WriteSignedFixed64(value)
	}
}

// WriteFloat32Slice writes a repeated float32 field to the buffer
func (w writer) WriteFloat32Slice(field int, values []float32) {
	w.Call("writePackedFloat", field, values)
}

// WriteFloat64Slice writes a repeated float64 field to the buffer
func (w writer) WriteFloat64Slice(field int, values []float64) {
	w.Call("writePackedDouble", field, values)
}

// WriteEnumSlice writes a repeated enum (as ints) to the buffer
func (w writer) WriteEnumSlice(field int, values []int) {
	w.Call("writePackedEnum", field, values)
}

// WriteBoolSlice writes a repeated bool field to the buffer
func (w writer) WriteBoolSlice(field int, values []bool) {
	w.Call("writePackedBool", field, values)
}
