package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/streamingriver/vitamins/apiserver"
	"github.com/streamingriver/vitamins/parser"
)

func main() {
	ctx := context.Background()

	// TestEventCreated("test", "alpine:latest")
	// return
	p := parser.New()

	p.Register("event-created", func(js string) {
		log.Printf("Got EventCreated")
		var ec EventCreated
		err := json.Unmarshal([]byte(js), &ec)
		if err != nil {
			log.Printf("Parse json failed with: %v", err)
			return
		}

		equals := false

		jss, err := inspect(ctx, ec.Name)
		if err != nil {
			log.Printf("%v", err)
		} else {
			equals = Equals(jss.Config.Env, ec.Env)
		}
		if equals == false {
			log.Printf("Deleteting container: %s", ec.Name)
			err := delete(ctx, ec.Name)
			if err != nil {
				log.Printf("%v", err)
			} else {
				log.Printf("Container deleted: %s", ec.Name)
			}
		}
		log.Printf("equals: %v", equals)
		if equals {
			// TODO: api call
			if ec.ServiceConfig != "" {
				buff := bytes.NewBuffer(nil)
				buff.Write([]byte("set-config"))
				buff.Write([]byte(" "))
				buff.Write([]byte(ec.ServiceConfig))
				http.DefaultClient.Post(fmt.Sprintf("http://%s:3080", ec.Name), "text/plain", buff)
			}
			log.Printf("container doesnt changed - nothing todo but,  maybe ping service API?")
			log.Printf("Calling api: http://%s:3080 with data: (%s)", ec.Name, encode(ec.ServiceConfig))
			return
		}
		log.Printf("Creating container: %s", ec.Name)

		err = create(
			ctx,              // context
			ec.Name,          // container name
			ec.Image,         // image url
			ec.Ports,         // ports mapping
			ec.Env,           // environment variables
			ec.Labels,        // labels
			ec.Cmd,           // command if needed
			ec.Auth.Username, // username if needed
			ec.Auth.Password, // password if needed
		)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		log.Printf("Container %s created", ec.Name)
	})

	p.Register("event-deleted", func(name string) {
		log.Printf("Deleteting container: %s", name)
		err := delete(ctx, name)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		log.Printf("Container %s deleted", name)
	})

	apiserver := apiserver.NewDefault(p)
	apiserver.Listen()

	runtime.Goexit()
}

// listContainers(ctx)
