package render

import (
	"fmt"
	"io"
)

type Content struct {
	writer io.Writer
	err    error
}

func NewContent(writer io.Writer) *Content {
	return &Content{writer: writer}
}

func (c *Content) WriteString(text string) {
	if c.err != nil {
		return
	}
	_, c.err = io.WriteString(c.writer, text)
}

func (c *Content) Writef(format string, args ...any) {
	if c.err != nil {
		return
	}
	_, c.err = fmt.Fprintf(c.writer, format, args...)
}

func (c *Content) Err() error {
	return c.err
}
