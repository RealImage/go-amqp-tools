package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jessevdk/go-flags"
	"github.com/streadway/amqp"
)

var opts struct {
	URL           string `short:"u" long:"url" description:"the AMQP URL to connect to"`
	Server        string `short:"s" long:"server" description:"the AMQP server to connect to"`
	Port          string `long:"port" description:"the port to connect on" default:"5672" `
	VHost         string `long:"vhost" description:"the vhost to use when connecting" default:"/"`
	Username      string `long:"username" description:"the username to login with" default:"guest"`
	Password      string `long:"password" description:"the password to login with" default:"guest"`
	Queue         string `short:"q" long:"queue" description:"the queue to consume from"`
	Exchange      string `short:"e" long:"exchange" description:"bind the queue to this exchange"`
	RoutingKey    string `short:"r" long:"routing-key" description:"the routing key to bind with"`
	Declare       bool   `short:"d" long:"declare" description:"declare an exclusive queue (deprecated, use --exclusive instead)"`
	Exclusive     bool   `short:"x" long:"exclusive" description:"declare the queue as exclusive"`
	NoAck         bool   `short:"A" long:"no-ack" description:"consume in no-ack mode"`
	Count         int    `short:"c" long:"count" description:"stop consuming after this many messages are consumed"`
	PrefetchCount int    `short:"p" long:"prefetch-count" description:"receive only this many message at a time from the server"`
}

func main() {
	optsParser := flags.NewParser(&opts, flags.Default)
	cmdArgs, err := optsParser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "parsing options failed: %s\n", err)
			os.Exit(1)
		}
	}

	if opts.URL == "" && opts.Server == "" {
		fmt.Fprintf(os.Stderr, "AMQP URL/server not specified\n")
		optsParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	if len(cmdArgs) == 0 {
		fmt.Fprintf(os.Stderr, "consuming command not specified\n")
		os.Exit(1)
	}

	if opts.Exchange == "" && opts.RoutingKey != "" {
		fmt.Fprintf(os.Stderr, "--routing-key option requires an exchange name to be provided with --exchange\n")
		os.Exit(1)
	}

	url := opts.URL
	if url == "" {
		url = fmt.Sprintf("amqp://%s:%s@%s:%d/%s", opts.Username, opts.Password, opts.Server, opts.Port, opts.VHost)
	}

	conn, err := amqp.Dial(opts.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dail failed: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	chann, err := conn.Channel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "channel creation failed: %s\n", err)
		os.Exit(1)
	}
	defer chann.Close()

	if opts.PrefetchCount != 0 {
		err = chann.Qos(opts.PrefetchCount, 0, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "channel qos failed: %s\n", err)
			os.Exit(1)
		}
	}

	if opts.Queue == "" {
		queue, err := chann.QueueDeclare(
			opts.Queue,     // queue
			true,           // durable
			true,           // autoDelete
			opts.Exclusive, // exclusive
			false,          // noWait
			nil,            // args
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "queue declare failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Server provided queue name: %s\n", queue.Name)
		opts.Queue = queue.Name

		if opts.Exchange != "" {
			err = chann.QueueBind(opts.Queue, opts.RoutingKey, opts.Exchange, false, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "queue bind failed: %s\n", err)
				os.Exit(1)
			}
		}
	}

	deliveries, err := chann.Consume(
		opts.Queue, // queue
		"",         // consumerTag
		opts.NoAck, // autoAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // args
	)
	if err != nil {
		fmt.Printf("consume failed: %s\n", err)
		os.Exit(1)
	}

	count := 0
	for d := range deliveries {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

		w, err := cmd.StdinPipe()
		if err != nil {
			fmt.Printf("retrieving cmd stdin failed: %s\n", err)
			os.Exit(1)
		}

		go func() {
			_, err := w.Write(d.Body)
			if err != nil {
				fmt.Printf("writing to cmd stdin failed: %s\n", err)
				os.Exit(1)
			}
			w.Close()
		}()

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("command execution failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Print(string(output))
		count++

		if !opts.NoAck {
			d.Ack(false)
		}

		if opts.Count != 0 && count >= opts.Count {
		}
	}
}
