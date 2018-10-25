package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/savaki/consulapi"
	"github.com/savaki/consulapi-example/example"
	"github.com/savaki/consulapi/connect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		name     = flag.String("name", "example-service", "name of service")
		n        = flag.Int("n", 1, "number of iterations to execute")
		api      = consulapi.NewHealth()
		resolver = connect.NewResolver(api, *name)
	)
	flag.Parse()

	conn, err := grpc.Dial("",
		grpc.WithBalancer(grpc.RoundRobin(resolver)),
		grpc.WithInsecure(),
	)
	check(err)
	defer conn.Close()

loop:
	for {
		state := conn.GetState()
		fmt.Println("state ->", state)
		switch state {
		case connectivity.Idle:
			conn.WaitForStateChange(ctx, state)
		case connectivity.Connecting:
			conn.WaitForStateChange(ctx, state)
		case connectivity.TransientFailure:
			return
		case connectivity.Shutdown:
			return
		case connectivity.Ready:
			break loop
		}
	}

	client := example.NewGreeterClient(conn)

	for i := 0; i < *n; i++ {
		out, err := client.Greet(ctx, &example.Input{Name: "world"})
		check(err)
		fmt.Println(out)
	}
}
