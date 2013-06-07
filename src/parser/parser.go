package parser

import (
	"bytes"
	"store/com"
	"strconv"
)

type CommandParser struct {
	count           int
	cur_count       int
	args            []string
	arg_len         int
	buffer          bytes.Buffer
	raw_buffer      bytes.Buffer
	commands        []com.Command
	skip            bool
	count_or_millis bool
}

func (c *CommandParser) ParseBytes(in_bytes []byte) (bool, []com.Command, error) {
	for _, v := range in_bytes {
		c.raw_buffer.WriteByte(v)
		if c.skip && v == '\n' {
			c.skip = false
			continue
		}
		if c.count == -1 {
			if v == '\r' {
				var err error
				if err != nil {
					panic(err)
				}
				count, err := strconv.Atoi(c.buffer.String())
				if err != nil {
					c.buffer.Reset()
					c.raw_buffer.Reset()
					continue
				}
				if c.count_or_millis {
					c.count = count
					c.buffer.Reset()
					c.args = []string{}
					c.arg_len = -1
					c.cur_count = 0
					c.skip = true
				} else {
					c.buffer.Reset()
					c.raw_buffer.Reset()
				}
			} else if v == '*' {
				c.count_or_millis = true
				c.raw_buffer.Reset()
				c.raw_buffer.WriteByte(v)
			} else if v == '&' {
				c.count_or_millis = false
				c.raw_buffer.Reset()
				c.raw_buffer.WriteByte(v)
			} else if v == '\n' {
				c.raw_buffer.Reset()
			} else {
				c.buffer.WriteByte(v)
			}
		} else {
			if c.arg_len == -1 {
				if v == '\r' {
					c.arg_len, _ = strconv.Atoi(c.buffer.String())
					c.buffer.Reset()
					c.skip = true
				} else if v != '$' {
					c.buffer.WriteByte(v)
				}
			} else {
				c.arg_len--
				if v == '\r' {
					c.args = append(c.args, c.buffer.String())
					c.buffer.Reset()
					c.cur_count += 1
					if c.cur_count == c.count {
						c.raw_buffer.WriteByte('\n')
						c.commands = append(c.commands, com.Command{c.args, c.raw_buffer.String(), nil})
						c.Restart()
					}
					c.skip = true
				} else {
					c.buffer.WriteByte(v)
				}
			}
		}
	}
	if len(c.commands) > 0 {
		return true, c.commands, nil
	} else {
		return false, nil, nil
	}
}

func (c *CommandParser) Reset() {
	c.commands = []com.Command{}
}

func (c *CommandParser) Restart() {
	c.count = -1
	c.cur_count = 0
	c.arg_len = -1
	c.args = []string{}
	c.buffer.Reset()
	c.raw_buffer.Reset()
	c.skip = false
	c.count_or_millis = true
}
