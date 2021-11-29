package main

import (
	"encoding/json"
	"fmt"
)

type EventCreated struct {
	Name   string            `json:"name"`
	Image  string            `json:"image"`
	Ports  map[string]string `json:"ports,omitempty"`
	Env    []string          `json:"env,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
	Cmd    []string          `json:"cmd,omitempty"`
}

func TestEventCreated(name, image string) *EventCreated {
	ec := &EventCreated{}
	ec.Name = name
	ec.Image = image
	ec.Env = []string{"HELLO=WORLD"}
	ec.Ports = map[string]string{"8080": "8080"}
	ec.Cmd = []string{"echo", "hello from piace of shit"}

	b, _ := json.MarshalIndent(ec, "", " ")
	fmt.Printf("%s\n", b)
	return ec
}
