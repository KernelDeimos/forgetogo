package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"time"
)

//go:embed example_output.txt
var ExampleOutput []byte

func main() {
	exitChan := make(chan struct{})

	// Simulate input
	go func() {
		for {
			var input string
			fmt.Scanln(&input)
			if input == "stop" {
				exitChan <- struct{}{}
			}
			if input == "args" {
				fmt.Printf("%v\n", os.Args)
			}
			if input == "crash" {
				os.Exit(1)
			}
		}
	}()

	// Simulate output
	go func() {
		byteReader := bytes.NewReader(ExampleOutput)
		reader := bufio.NewReader(byteReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil && err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			fmt.Print(line)
			time.Sleep(time.Microsecond * 30)
		}

		for {
			fmt.Println("test output")
			<-time.After(time.Second)
		}
	}()

	<-exitChan
}
