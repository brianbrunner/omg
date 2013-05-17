package parser

import (
    "strconv"
    "bytes"
    "store/com"
)

type CommandParser struct {
    count int
    cur_count int
    args []string
    arg_len int
    buffer bytes.Buffer
    raw_buffer bytes.Buffer
    commands []com.Command
}

func (c *CommandParser) ParseBytes(in_bytes []byte) (bool, []com.Command, error) {
    skip := false
    for _, v := range in_bytes {
        c.raw_buffer.WriteByte(v)
        if skip && v == '\n' {
            skip = false
            continue
        }
        if c.count == -1 {
            if v == '\r' {
                c.count, _ = strconv.Atoi(c.buffer.String())
                c.buffer.Reset()
                c.args = []string{}
                c.arg_len = -1
                c.cur_count = 0
                skip = true
            } else if (v != '*') {
                c.buffer.WriteByte(v)
            }
        } else {
            if c.arg_len == -1 {
                if v == '\r' {
                    c.arg_len, _ = strconv.Atoi(c.buffer.String())
                    c.buffer.Reset()
                    skip = true
                } else if v != '$' {
                    c.buffer.WriteByte(v)
                }
            } else {
                c.arg_len--
                if v == '\r' {
                    c.args = append(c.args,c.buffer.String())
                    c.buffer.Reset()
                    c.cur_count += 1
                    if c.cur_count == c.count {
                        c.commands = append(c.commands,com.Command{c.args,c.raw_buffer.String(),nil})
                    }
                    skip = true
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
    c.count = -1
    c.cur_count = 0
    c.arg_len = -1
    c.args = []string{}
    c.buffer.Reset()
    c.raw_buffer.Reset()
    c.commands = []com.Command{}
}
