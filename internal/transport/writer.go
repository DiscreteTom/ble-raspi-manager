package transport

import (
	"bytes"

	"github.com/google/uuid"
)

type WriteHandler struct {
	uuid     []byte
	content  []byte
	callback func(uuid []byte, content []byte)
}

func NewWriteHandler(callback func(uuid []byte, content []byte)) *WriteHandler {
	return &WriteHandler{
		uuid:     []byte(uuid.Nil.String()),
		content:  []byte{},
		callback: callback,
	}
}

func (w *WriteHandler) Write(part []byte) {
	if !bytes.Equal(w.uuid, part[:len(w.uuid)]) {
		// mismatch uuid, new command
		w.uuid = part[:len(w.uuid)]
		w.content = []byte{}
	}

	if len(part) == len(w.uuid) {
		w.callback(w.uuid, w.content)
	} else {
		w.content = append(w.content, part[len(w.uuid):]...)
	}
}
