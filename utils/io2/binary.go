package io2

import (
	"bufio"
	"encoding/binary"
	"io"
)

// TODO2 定义BinaryWriter和BinaryReader类型，下面的函数作为它的成员函数

func WriteBool(writer *bufio.Writer, val bool) error {
	v := byte(0)
	if val {
		v = 1
	}

	if err := writer.WriteByte(v); err != nil {
		return err
	}
	return nil
}

func ReadBool(reader *bufio.Reader) (bool, error) {
	v, err := reader.ReadByte()
	if err != nil {
		return false, err
	}
	return v > 0, nil
}

func WriteUint8Field(writer *bufio.Writer, data uint8) error {
	if err := writer.WriteByte(data); err != nil {
		return err
	}
	return nil
}

func ReadUint8Field(reader *bufio.Reader) (uint8, error) {
	data, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return data, nil
}

func WriteUint16Field(writer *bufio.Writer, data uint16) error {
	dataBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(dataBytes, data)

	err := WriteAll(writer, dataBytes)
	if err != nil {
		return err
	}
	return nil
}

func ReadUint16Field(reader *bufio.Reader) (uint16, error) {
	dataBytes := make([]byte, 2)
	_, err := io.ReadFull(reader, dataBytes)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(dataBytes), nil
}

func WriteUint32Field(writer *bufio.Writer, data uint32) error {
	dataBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataBytes, data)

	err := WriteAll(writer, dataBytes)
	if err != nil {
		return err
	}
	return nil
}

func ReadUint32Field(reader *bufio.Reader) (uint32, error) {
	dataBytes := make([]byte, 4)
	_, err := io.ReadFull(reader, dataBytes)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(dataBytes), nil
}

func WriteUint64Field(writer *bufio.Writer, data uint64) error {
	dataBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataBytes, data)

	err := WriteAll(writer, dataBytes)
	if err != nil {
		return err
	}
	return nil
}

func ReadUint64Field(reader *bufio.Reader) (uint64, error) {
	dataBytes := make([]byte, 8)
	_, err := io.ReadFull(reader, dataBytes)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(dataBytes), nil
}

func WriteStringField(writer *bufio.Writer, data string) error {
	dataBytes := []byte(data)
	if err := writer.WriteByte(byte(len(dataBytes))); err != nil {
		return err
	}

	err := WriteAll(writer, dataBytes)
	if err != nil {
		return err
	}
	return nil
}

func ReadStringField(reader *bufio.Reader) (string, error) {
	length, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	dataBytes := make([]byte, length)
	_, err = io.ReadFull(reader, dataBytes)
	if err != nil {
		return "", err
	}
	return string(dataBytes), nil
}
