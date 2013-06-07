package com

type Command struct {
	Args       []string
	CommandRaw string
	ReplyChan  chan string
}
