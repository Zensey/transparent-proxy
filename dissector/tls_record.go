package dissector

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	RecordHeaderLen = 5
)

// record content type
const (
	ChangeCipherSpec = 0x14
	EncryptedAlert   = 0x15
	Handshake        = 0x16
	AppData          = 0x17
)

var (
	ErrBadType = errors.New("bad type")
)

type Version uint16

type Record struct {
	Type    uint8
	Version Version
	Length  int
}

func ReadRecord(r io.Reader) (*Record, error) {
	record := &Record{}
	if _, err := record.ReadFrom(r); err != nil {
		return nil, err
	}
	return record, nil
}

func (rec *Record) ReadFrom(r io.Reader) (n int64, err error) {
	b := make([]byte, RecordHeaderLen)
	nn, err := io.ReadFull(r, b)
	n += int64(nn)
	if err != nil {
		return
	}
	rec.Type = b[0]
	rec.Version = Version(binary.BigEndian.Uint16(b[1:3]))
	length := int(binary.BigEndian.Uint16(b[3:5]))
	rec.Length = length

	return
}

func (rec *Record) Valid() bool {
	if rec.Type != 0x16 {
		return false
	}
	return true
}
