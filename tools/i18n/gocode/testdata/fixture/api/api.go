package api

type Key string

type Message struct {
	Key Key
}

const KeyInternal Key = "request.internal"

func Fail(key Key) Message {
	return Message{Key: key}
}
