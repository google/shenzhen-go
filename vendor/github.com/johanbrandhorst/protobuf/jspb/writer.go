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

// NewWriter returns a new Writer ready for writing.
func NewWriter() Writer {
	return &writer{
		Object: js.Global.Get("BinaryWriter").New(),
	}
}

// writer implements the Writer interface.
type writer struct {
	*js.Object
}

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
	w.Call("writeInt64", field, value)
}

// WriteUint32 writes a uint32 field to the buffer.
func (w writer) WriteUint32(field int, value uint32) {
	w.Call("writeUint32", field, value)
}

// WriteUint64 writes a uint64 field to the buffer.
func (w writer) WriteUint64(field int, value uint64) {
	w.Call("writeUint64", field, value)
}

// WriteSint32 writes an sint32 field to the buffer.
func (w writer) WriteSint32(field int, value int32) {
	w.Call("writeSint32", field, value)
}

// WriteSint64 writes an sint64 field to the buffer.
func (w writer) WriteSint64(field int, value int64) {
	w.Call("writeSint64", field, value)
}

// WriteFixed32 writes a fixed32 field to the buffer.
func (w writer) WriteFixed32(field int, value uint32) {
	w.Call("writeFixed32", field, value)
}

// WriteFixed64 writes a fixed64 field to the buffer.
func (w writer) WriteFixed64(field int, value uint64) {
	w.Call("writeFixed64", field, value)
}

// WriteSfixed32 writes an sfixed32 field to the buffer.
func (w writer) WriteSfixed32(field int, value int32) {
	w.Call("writeSfixed32", field, value)
}

// WriteSfixed64 writes an sfixed64 field to the buffer.
func (w writer) WriteSfixed64(field int, value int64) {
	w.Call("writeSfixed64", field, value)
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
func (w writer) WriteInt32Slice(field int, value []int32) {
	w.Call("writePackedInt32", field, value)
}

// WriteInt64Slice writes a repeated int64 field to the buffer.
func (w writer) WriteInt64Slice(field int, value []int64) {
	w.Call("writePackedInt64", field, value)
}

// WriteUint32Slice writes a repeated uint32 field to the buffer.
func (w writer) WriteUint32Slice(field int, value []uint32) {
	w.Call("writePackedUint32", field, value)
}

// WriteUint64Slice writes a repeated uint64 field to the buffer.
func (w writer) WriteUint64Slice(field int, value []uint64) {
	w.Call("writePackedUint64", field, value)
}

// WriteSint32Slice writes a repeated sint32 field to the buffer.
func (w writer) WriteSint32Slice(field int, value []int32) {
	w.Call("writePackedSint32", field, value)
}

// WriteSint64Slice writes a repeated sint64 field to the buffer.
func (w writer) WriteSint64Slice(field int, value []int64) {
	w.Call("writePackedSint64", field, value)
}

// WriteFixed32Slice writes a repeated fixed32 field to the buffer.
func (w writer) WriteFixed32Slice(field int, value []uint32) {
	w.Call("writePackedFixed32", field, value)
}

// WriteFixed64Slice writes a repeated fixed64 field to the buffer.
func (w writer) WriteFixed64Slice(field int, value []uint64) {
	w.Call("writePackedFixed64", field, value)
}

// WriteSfixed32Slice writes a repeated sfixed32 field to the buffer.
func (w writer) WriteSfixed32Slice(field int, value []int32) {
	w.Call("writePackedSfixed32", field, value)
}

// WriteSfixed64Slice writes a repeated sfixed64 field to the buffer.
func (w writer) WriteSfixed64Slice(field int, value []int64) {
	w.Call("writePackedSfixed64", field, value)
}

// WriteFloat32Slice writes a repeated float32 field to the buffer
func (w writer) WriteFloat32Slice(field int, value []float32) {
	w.Call("writePackedFloat", field, value)
}

// WriteFloat64Slice writes a repeated float64 field to the buffer
func (w writer) WriteFloat64Slice(field int, value []float64) {
	w.Call("writePackedDouble", field, value)
}

// WriteEnumSlice writes a repeated enum (as ints) to the buffer
func (w writer) WriteEnumSlice(field int, value []int) {
	w.Call("writePackedEnum", field, value)
}

// WriteBoolSlice writes a repeated bool field to the buffer
func (w writer) WriteBoolSlice(field int, value []bool) {
	w.Call("writePackedBool", field, value)
}
