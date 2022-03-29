package transport

type ReadHandler struct {
	uuid    []byte
	content []byte
	index   int
}

func NewReadHandler(uuid, content []byte) *ReadHandler {
	return &ReadHandler{
		uuid:    uuid,
		content: content,
		index:   0,
	}
}

func (r *ReadHandler) Read() []byte {
	if r.index < len(r.content) {
		end := r.index + 256 // send 256 byte content everytime
		if end > len(r.content) {
			end = len(r.content)
		}
		result := append(r.uuid, r.content[r.index:end]...)
		r.index = end
		return result
	}

	return r.uuid // if read finish, return uuid only
}
