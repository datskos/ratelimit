package storage

import (
	"bytes"
	"encoding/gob"
)

func decode(data []byte) (*Value, error) {
	var value Value
	buffer := bytes.NewReader(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(&value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func encode(value *Value) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(&value)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
