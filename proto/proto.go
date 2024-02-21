package proto

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"io"
	"through/log"
)

// ReadMeta read data from reader and unmarshal
func ReadMeta(reader io.Reader) (meta *Meta, err error) {
	// read data length
	header := make([]byte, 4)
	if _, err = io.ReadFull(reader, header); err != nil {
		log.Error("reader header error: %v", err)
		return
	}
	dataLen := binary.BigEndian.Uint32(header)

	buf := make([]byte, dataLen)
	if _, err = io.ReadFull(reader, buf); err != nil {
		log.Error("reader data error: %v", err)
		return
	}

	meta = &Meta{}
	err = proto.Unmarshal(buf, meta)

	return
}

// WriteMeta marshal meta and write to writer
func WriteMeta(writer io.Writer, meta *Meta) (err error) {
	var data []byte
	if data, err = proto.Marshal(meta); err != nil {
		return
	}
	dataLen := uint32(len(data))

	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, dataLen)

	_, err = writer.Write(append(header, data...))
	return
}
