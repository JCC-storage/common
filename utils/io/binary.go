package io

import (
	"bufio"
	"encoding/binary"
	"io"
)

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
