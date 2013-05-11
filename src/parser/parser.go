package parser

import (
    "strconv"
    "bytes"
)

type CommandParser struct {
    command_length int
    current_position int
    command_raw bytes.Buffer
    command_parts []string
}

func (c *CommandParser) AddArg(arg string) (done bool, command []string, commandRaw string, err error) {
    c.command_raw.WriteString(arg)
    arg = arg[0:len(arg)-2]
    if c.command_length == 0 {
        length, err := strconv.Atoi(arg[1:len(arg)])
        if err != nil {
            c.reset()
            return true, []string{"-ERR"}, "", err
        }
        c.command_length = length
        c.command_parts = make([]string,length)
    } else if c.current_position % 2 == 0 {
        cur_i := c.current_position/2-1
        c.command_parts[cur_i] = arg
        if cur_i+1 == c.command_length {
            command := c.command_parts
            command_raw := c.command_raw.String()
            c.reset()
            return true, command, command_raw, nil
        }
    }
    c.current_position += 1
    return false, []string{""}, "", nil
}

func (c *CommandParser) reset() {
    c.command_length = 0
    c.current_position = 0
    c.command_raw.Reset()
}
