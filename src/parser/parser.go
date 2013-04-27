package parser

import (
    "strconv"
    "strings"
    "bytes"
)

type CommandParser struct {
    command_length int
    current_position uint8
    command_raw bytes.Buffer
    command_parts []string
}

func (c *CommandParser) AddArg(arg string) (done bool, command []string, commandRaw string, err error) {
    c.command_raw.WriteString(arg)
    arg = strings.TrimSpace(arg)
    if c.command_length == 0 {
        length, err := strconv.Atoi(arg[1:len(arg)])
        if err != nil {
            c.reset()
            return true, []string{"-ERR"}, "", err
        }
        c.command_length = length
    } else if c.current_position % 2 == 0 {
        c.command_parts = append(c.command_parts,arg)
        l := len(c.command_parts)
        if l == c.command_length {
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
    c.command_parts = make([]string,0)
    c.command_raw.Reset()
}
