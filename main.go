package main

import (
	"context"
	"encoding/json"
	"log"
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
			log.Printf("container doesnt changed - nothing todo but,  maybe ping service API?")
			return
		}
		log.Printf("Creating container: %s", ec.Name)

		err = create(
			ctx,       // context
			ec.Name,   // container name
			ec.Image,  // image url
			ec.Ports,  // ports mapping
			ec.Env,    // environment variables
			ec.Labels, // labels
			ec.Cmd,    // command if needed
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
