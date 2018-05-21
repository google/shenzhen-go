package jspb

import "github.com/gopherjs/gopherjs/js"

// Reader encapsulates the jspb.BinaryReader.
type Reader interface {
	Next() bool
	Err() error

	GetFieldNumber() int
	SkipField()

	// Scalars
	ReadInt32() int32
	ReadInt64() int64
	ReadUint32() uint32
	ReadUint64() uint64
	ReadSint32() int32
	ReadSint64() int64
	ReadFixed32() uint32
	ReadFixed64() uint64
	ReadSfixed32() int32
	ReadSfixed64() int64
	ReadFloat32() float32
	ReadFloat64() float64
	ReadEnum() int
	ReadBool() bool
	ReadString() string
	ReadBytes() []byte

	// Scalar Slices
	ReadInt32Slice() []int32
	ReadInt64Slice() []int64
	ReadUint32Slice() []uint32
	ReadUint64Slice() []uint64
	ReadSint32Slice() []int32
	ReadSint64Slice() []int64
	ReadFixed32Slice() []uint32
	ReadFixed64Slice() []uint64
	ReadSfixed32Slice() []int32
	ReadSfixed64Slice() []int64
	ReadFloat32Slice() []float32
	ReadFloat64Slice() []float64
	ReadEnumSlice() []int
	ReadBoolSlice() []bool

	// Specials
	ReadMessage(func())
}

type decoder struct {
	*js.Object
}

// ReadInt64 reads a signed 64 bit varint
func (d decoder) ReadInt64() int64 {
	d.Call("readSplitVarint64_")
	high := d.Get("tempHigh_").Uint64()
	low := d.Get("tempLow_").Uint64()
	return d.mergeSignedInteger(low, high)
}

// ReadZigzag64 reads a signed 64 bit Zigzag encoded varint
func (d decoder) ReadZigzag64() int64 {
	v := d.ReadUint64()
	// https://github.com/gogo/protobuf/blob/1ef32a8b9fc3f8ec940126907cedb5998f6318e4/proto/decode.go#L254
	return int64((v >> 1) ^ uint64((int64(v&1)<<63)>>63))
}

// ReadUint64 reads a unsigned 64 bit varint
func (d decoder) ReadUint64() uint64 {
	d.Call("readSplitVarint64_")
	high := d.Get("tempHigh_").Uint64()
	low := d.Get("tempLow_").Uint64()
	return high<<32 | low
}

// ReadFixed64 reads a 64 bit unsigned integer
func (d decoder) ReadFixed64() uint64 {
	low := d.Call("readUint32").Uint64()
	high := d.Call("readUint32").Uint64()
	return high<<32 | low
}

// ReadSignedFixed64 reads a 64 bit signed integer
func (d decoder) ReadSignedFixed64() int64 {
	low := d.Call("readUint32").Uint64()
	high := d.Call("readUint32").Uint64()
	return d.mergeSignedInteger(low, high)
}

func (d decoder) mergeSignedInteger(low, high uint64) int64 {
	// Adapted from google-protobuf
	// https://github.com/google/protobuf/blob/25625b956a2f0d432582009c16553a9fd21c3cea/js/binary/utils.js#L521
	negative := (high&0x8 == 1)
	if negative {
		low = (^low + 1)
		var carry uint64
		if low == 0 {
			carry = 1
		}
		high = (^high + carry)
	}
	return int64(high<<32 | low)
}

// NewReader returns a new Reader ready for writing
func NewReader(data []byte) Reader {
	r := &reader{
		Object: js.Global.Get("BinaryReader").New(data),
	}
	r.decoder = &decoder{
		Object: r.Get("decoder_"),
	}
	return r
}

// reader implements the Reader interface
type reader struct {
	*js.Object
	decoder *decoder
	err     error
}

// Reads the next field header in the stream if there is one, returns true if
// we saw a valid field header or false if we've read the whole stream.
// Sets err if we encountered a deprecated START_GROUP/END_GROUP fielr.decoder.
func (r *reader) Next() bool {
	defer catchException(&r.err)
	return r.err == nil && r.Call("nextField").Bool() && !r.Call("isEndGroup").Bool()
}

// Err returns the error state of the Reader. It must
// be called after Next() has returned false.
func (r reader) Err() error {
	return r.err
}

// The field number of the next field in the buffer, or
// -1 if there is no next fielr.decoder.
func (r reader) GetFieldNumber() int {
	return r.Call("getFieldNumber").Int()
}

// Skips over the next field in the binary stream.
func (r reader) SkipField() {
	r.Call("skipField")
}

// ReadInt32 reads a signed 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadInt32() int32 {
	defer catchException(&r.err)
	return int32(r.Call("readInt32").Int())
}

// ReadInt64 reads a signed 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadInt64() int64 {
	defer catchException(&r.err)
	return r.decoder.ReadInt64()
}

// ReadUint32 reads an unsigned 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadUint32() uint32 {
	defer catchException(&r.err)
	return uint32(r.Call("readUint32").Int())
}

// ReadUint64 reads an unsigned 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadUint64() uint64 {
	defer catchException(&r.err)
	return r.decoder.ReadUint64()
}

// ReadSint32 reads a signed 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadSint32() int32 {
	defer catchException(&r.err)
	return int32(r.Call("readSint32").Int())
}

// ReadSint64 reads a signed 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadSint64() int64 {
	defer catchException(&r.err)
	return r.decoder.ReadZigzag64()
}

// ReadFixed32 reads an unsigned 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadFixed32() uint32 {
	defer catchException(&r.err)
	return uint32(r.Call("readFixed32").Int())
}

// ReadFixed64 reads an unsigned 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadFixed64() uint64 {
	defer catchException(&r.err)
	return r.decoder.ReadFixed64()
}

// ReadSfixed32 reads a signed 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadSfixed32() int32 {
	defer catchException(&r.err)
	return int32(r.Call("readSfixed32").Int())
}

// ReadSfixed64 reads a signed 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadSfixed64() int64 {
	defer catchException(&r.err)
	return r.decoder.ReadSignedFixed64()
}

// ReadFloat32 reads a 32-bit floating point field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadFloat32() float32 {
	defer catchException(&r.err)
	return float32(r.Call("readFloat").Float())
}

// ReadFloat64 reads a 64-bit floating point field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadFloat64() float64 {
	defer catchException(&r.err)
	return r.Call("readDouble").Float()
}

// ReadEnum reads an enum field from the binary stream,
// or sets err if the next field in the stream
// is not of the correct wire type.
func (r *reader) ReadEnum() int {
	defer catchException(&r.err)
	return r.Call("readEnum").Int()
}

// ReadBool reads a bool field from the binary stream, or sets err
// if the next field in the stream is not of the correct wire type.
func (r *reader) ReadBool() bool {
	defer catchException(&r.err)
	return r.Call("readBool").Bool()
}

// ReadString reads a string field from the binary stream, or sets err
// if the next field in the stream is not of the correct wire type.
func (r *reader) ReadString() string {
	defer catchException(&r.err)
	return r.Call("readString").String()
}

// ReadBytes reads a bytes field from the binary stream, or sets err
// if the next field in the stream is not of the correct wire type.
func (r *reader) ReadBytes() []byte {
	defer catchException(&r.err)
	return r.Call("readBytes").Interface().([]byte)
}

// ReadMessage deserializes a proto using
// the provided reader function.
func (r *reader) ReadMessage(readFunc func()) {
	defer catchException(&r.err)
	r.Call("readMessage", js.Undefined /* Unused */, readFunc)
}

// ReadInt32Slice reads a repeated signed 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadInt32Slice() (ret []int32) {
	defer catchException(&r.err)
	values := r.Call("readPackedInt32").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, int32(value.(float64)))
	}

	return ret
}

// ReadInt64Slice reads a repeated signed 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadInt64Slice() (ret []int64) {
	defer catchException(&r.err)
	cb := js.MakeFunc(func(*js.Object, []*js.Object) interface{} {
		ret = append(ret, r.decoder.ReadInt64())
		return js.Undefined
	})
	r.Call("readPackedField_", cb)

	return ret
}

// ReadUint32Slice reads a repeated unsigned 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadUint32Slice() (ret []uint32) {
	defer catchException(&r.err)
	values := r.Call("readPackedUint32").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, uint32(value.(float64)))
	}

	return ret
}

// ReadUint64Slice reads a repeated unsigned 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadUint64Slice() (ret []uint64) {
	defer catchException(&r.err)
	cb := js.MakeFunc(func(*js.Object, []*js.Object) interface{} {
		ret = append(ret, r.decoder.ReadUint64())
		return js.Undefined
	})
	r.Call("readPackedField_", cb)

	return ret
}

// ReadSint32Slice reads a repeated signed 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadSint32Slice() (ret []int32) {
	defer catchException(&r.err)
	values := r.Call("readPackedSint32").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, int32(value.(float64)))
	}

	return ret
}

// ReadSint64Slice reads a repeated signed 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadSint64Slice() (ret []int64) {
	defer catchException(&r.err)
	cb := js.MakeFunc(func(*js.Object, []*js.Object) interface{} {
		ret = append(ret, r.decoder.ReadZigzag64())
		return js.Undefined
	})
	r.Call("readPackedField_", cb)

	return ret
}

// ReadFixed32Slice reads a repeated unsigned 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadFixed32Slice() (ret []uint32) {
	defer catchException(&r.err)
	values := r.Call("readPackedFixed32").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, uint32(value.(float64)))
	}

	return ret
}

// ReadFixed64Slice reads a repeated unsigned 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadFixed64Slice() (ret []uint64) {
	defer catchException(&r.err)
	cb := js.MakeFunc(func(*js.Object, []*js.Object) interface{} {
		ret = append(ret, r.decoder.ReadFixed64())
		return js.Undefined
	})
	r.Call("readPackedField_", cb)

	return ret
}

// ReadSfixed32Slice reads a repeated signed 32-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadSfixed32Slice() (ret []int32) {
	defer catchException(&r.err)
	values := r.Call("readPackedSfixed32").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, int32(value.(float64)))
	}

	return ret
}

// ReadSfixed64Slice reads a repeated signed 64-bit integer field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadSfixed64Slice() (ret []int64) {
	defer catchException(&r.err)
	cb := js.MakeFunc(func(*js.Object, []*js.Object) interface{} {
		ret = append(ret, r.decoder.ReadSignedFixed64())
		return js.Undefined
	})
	r.Call("readPackedField_", cb)

	return ret
}

// ReadFloat32Slice reads a repeated 32-bit floating point field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadFloat32Slice() (ret []float32) {
	defer catchException(&r.err)
	values := r.Call("readPackedFloat").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, float32(value.(float64)))
	}

	return ret
}

// ReadFloat64Slice reads a repeated 64-bit floating point field from the binary
// stream, or sets err if the next field in the
// stream is not of the correct wire type.
func (r *reader) ReadFloat64Slice() (ret []float64) {
	defer catchException(&r.err)
	values := r.Call("readPackedDouble").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, value.(float64))
	}

	return ret
}

// ReadEnumSlice reads a repeated enum field from the binary stream,
// or sets err if the next field in the stream
// is not of the correct wire type.
func (r *reader) ReadEnumSlice() (ret []int) {
	defer catchException(&r.err)
	values := r.Call("readPackedEnum").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, int(value.(float64)))
	}

	return ret
}

// ReadBoolSlice reads a repeated bool field from the binary stream, or sets err
// if the next field in the stream is not of the correct wire type.
func (r *reader) ReadBoolSlice() (ret []bool) {
	defer catchException(&r.err)
	values := r.Call("readPackedBool").Interface().([]interface{})
	for _, value := range values {
		ret = append(ret, value.(bool))
	}

	return ret
}
