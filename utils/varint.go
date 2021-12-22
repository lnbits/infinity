package utils

// this is compatible with embit's "compact" (also used in bitcoind)
// https://github.com/diybitcoinhardware/embit/blob/master/src/embit/compact.py

import (
	"encoding/binary"
	"errors"
	"io"
)

var ErrVarIntNotCanonical = errors.New("decoded varint is not minimally encoded")

func ReadVarInt(r io.Reader) (uint64, error) {
	var buf [8]byte
	_, err := io.ReadFull(r, buf[:1])
	if err != nil {
		return 0, err
	}
	discriminant := buf[0]

	var rv uint64
	switch {
	case discriminant < 0xfd:
		rv = uint64(discriminant)

	case discriminant == 0xfd:
		_, err := io.ReadFull(r, buf[:2])
		switch {
		case err == io.EOF:
			return 0, io.ErrUnexpectedEOF
		case err != nil:
			return 0, err
		}
		rv = uint64(binary.LittleEndian.Uint16(buf[:2]))

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		if rv < 0xfd {
			return 0, ErrVarIntNotCanonical
		}

	case discriminant == 0xfe:
		_, err := io.ReadFull(r, buf[:4])
		switch {
		case err == io.EOF:
			return 0, io.ErrUnexpectedEOF
		case err != nil:
			return 0, err
		}
		rv = uint64(binary.LittleEndian.Uint32(buf[:4]))

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		if rv <= 0xffff {
			return 0, ErrVarIntNotCanonical
		}

	default:
		_, err := io.ReadFull(r, buf[:])
		switch {
		case err == io.EOF:
			return 0, io.ErrUnexpectedEOF
		case err != nil:
			return 0, err
		}
		rv = binary.LittleEndian.Uint64(buf[:])

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		if rv <= 0xffffffff {
			return 0, ErrVarIntNotCanonical
		}
	}

	return rv, nil
}

func WriteVarInt(w io.Writer, val uint64) error {
	var buf [8]byte
	var length int
	switch {
	case val < 0xfd:
		buf[0] = uint8(val)
		length = 1

	case val <= 0xffff:
		buf[0] = uint8(0xfd)
		binary.LittleEndian.PutUint16(buf[1:3], uint16(val))
		length = 3

	case val <= 0xffffffff:
		buf[0] = uint8(0xfe)
		binary.LittleEndian.PutUint32(buf[1:5], uint32(val))
		length = 5

	default:
		buf[0] = uint8(0xff)
		_, err := w.Write(buf[:1])
		if err != nil {
			return err
		}

		binary.LittleEndian.PutUint64(buf[:], uint64(val))
		length = 8
	}

	_, err := w.Write(buf[:length])
	return err
}
