package reply

import (
	"bytes"
	"fmt"
)

var OKReply string = "+OK\r\n"

var NilReply string = "$-1\r\n"

func ErrorReply(error interface{}) string {
	return fmt.Sprintf("-ERR %s\r\n", error)
}

func IntReply(number int) string {
	return fmt.Sprintf(":%d\r\n", number)
}

func BulkReply(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}

type MultiBulkWriter struct {
	reply bytes.Buffer
}

func (w *MultiBulkWriter) WriteCount(number int) {
	w.reply.WriteString(fmt.Sprintf("*%d\r\n", number))
}

func (w *MultiBulkWriter) WriteString(str string) {
	w.reply.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
}

func (w *MultiBulkWriter) WriteNil() {
	w.reply.WriteString(NilReply)
}

func (w *MultiBulkWriter) String() string {
	return w.reply.String()
}
