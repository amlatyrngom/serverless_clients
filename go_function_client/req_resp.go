package function_client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type FunctionState int

const (
	QUEUED      FunctionState = iota
	RUNNING                   = iota
	NOTDEPLOYED               = iota
	INTERRUPTED               = iota
	MISSING                   = iota
)

type InvokeReq struct {
	FunctionID int64 `json:"function_id"`
}

type InvokeResp struct {
	InvocationID  int64 `json:"invocation_id"`
	FunctionState `json:"function_state"`
}

type InterruptReq struct {
	InvocationID int64 `json:"invocation_id"`
}

type InterruptResp struct {
	InvocationID  int64 `json:"invocation_id"`
	FunctionState `json:"function_state"`
}

type FindFuncReq struct {
	FunctionID int64 `json:"function_id"`
}

type FindFuncResp struct {
	FunctionState `json:"function_state"`
}

const SIZE64 uint64 = 8

func DecodeBytes(r io.Reader) ([][]byte, error) {
	int_buffer := make([]byte, SIZE64)
	_, err := io.ReadFull(r, int_buffer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	num_elems := binary.LittleEndian.Uint64(int_buffer)
	elem_sizes := make([]uint64, num_elems)
	for i := uint64(0); i < num_elems; i++ {
		_, err := io.ReadFull(r, int_buffer)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		elem_size := binary.LittleEndian.Uint64(int_buffer)
		elem_sizes[i] = elem_size
	}
	elems := make([][]byte, num_elems)
	for i := uint64(0); i < num_elems; i++ {
		elems[i] = make([]byte, elem_sizes[i])
		_, err = io.ReadFull(r, elems[i])
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	return elems, nil
}

func EncodeBytes(elems [][]byte) ([]byte, error) {
	num_elems := uint64(len(elems))
	elem_sizes := make([]uint64, num_elems)
	for i, elem := range elems {
		elem_sizes[i] = uint64(len(elem))
	}
	response := new(bytes.Buffer)
	err := binary.Write(response, binary.LittleEndian, num_elems)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return nil, err
	}
	for _, resp_size := range elem_sizes {
		err := binary.Write(response, binary.LittleEndian, resp_size)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
			return nil, err
		}
	}
	for _, resp := range elems {
		response.Write(resp)
	}
	return response.Bytes(), nil
}
