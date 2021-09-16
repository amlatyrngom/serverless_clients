package function_client

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvokeArgs(t *testing.T) {
	http_client := &http.Client{
		Timeout: time.Second * 10,
	}
	args := [][]byte{[]byte("Arg1"), []byte("Arg2"), []byte("Arg333333333333333")}
	req, err := EncodeBytes(args)
	assert.Equal(t, err, nil, "Got Error")
	r, err := http_client.Post("http://localhost:37000/invoke", "application/octet-stream", bytes.NewBuffer(req))
	assert.Equal(t, err, nil, "Got Error")
	assert.Equal(t, r.StatusCode, http.StatusOK, "Got Bad Code")
	resp, err := DecodeBytes(r.Body)
	assert.Equal(t, err, nil, "Got Error")
	for _, arg := range resp {
		fmt.Println("Got response", string(arg))
	}
}
