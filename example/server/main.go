package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/savaki/consulapi-example/example"
	"github.com/savaki/consulapi/connect"
	"google.golang.org/grpc"
)

type greeter struct {
}

func (g greeter) Greet(ctx context.Context, in *example.Input) (*example.Output, error) {
	fmt.Printf("Greet(%v)\n", in.Name)
	return &example.Output{
		Greeting: "Hello " + in.Name,
	}, nil
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	var (
		port = flag.Int("port", 8080, "port to listen on")
		name = flag.String("name", "example-service", "name of service")
	)
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", *port))
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	server := grpc.NewServer()
	defer server.Stop()

	example.RegisterGreeterServer(server, greeter{})

	go server.Serve(listener)

	fn := func() error {
		fmt.Println("health check")
		return nil
	}
	closer, err := connect.Register(*name, *port, connect.WithHealthCheckFunc(fn))
	check(err)
	defer closer.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop
}
