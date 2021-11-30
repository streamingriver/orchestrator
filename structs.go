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
	Auth   struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"auth"`

	ServiceConfig string `json:"serviceConfig,omitempty`
}

func TestEventCreated(name, image string) *EventCreated {
	ec := &EventCreated{}
	ec.Name = name
	ec.Image = image
	ec.Env = []string{"HELLO=WORLD"}
	ec.Ports = map[string]string{"127.0.0.1:8080": "8080/udp"}
	ec.Cmd = []string{"echo", "hello from other side"}
	ec.Auth.Username = "user"
	ec.Auth.Password = "password"
	ec.ServiceConfig = fmt.Sprintf("set-config %s", "[base64encoded_data]")

	b, _ := json.MarshalIndent(ec, "", " ")
	fmt.Printf("%s\n", b)
	return ec
}
