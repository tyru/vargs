package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	nulTerminated = flag.Bool("0", false, "Change separator to NUL character. This is same as \"-s nul\"")
	separators    = flag.String("s", "newline", "Change separators with these comma-separated values (available values are \"space\", \"tab\", \"newline\", \"nul\")")
	replaceStr    = flag.String("I", "", "If this replacement string was given, replace arguments by this with each item")
)

func main() {
	flag.Parse()

	if *nulTerminated {
		*separators = "nul"
	}

	var msg []string
	if flag.NArg() == 0 {
		if *replaceStr == "" {
			msg = []string{"drop"}
		} else {
			msg = []string{"drop", *replaceStr}
		}
	} else {
		msg = flag.Args()
	}
	existsReplaceStr := false
	for _, m := range msg {
		if strings.Contains(m, *replaceStr) {
			existsReplaceStr = true
			break
		}
	}
	if !existsReplaceStr {
		fmt.Fprintln(os.Stderr, "warning: -I {replstr} option was specified but no {replstr} in arguments")
	}
	buildMsg := makeMsgBuilder(msg)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan string)
	done := make(chan bool)
	var senderErr error
	go func() {
		for {
			select {
			case item, open := <-ch:
				if !open {
					done <- true
					return
				}
				newMsg := buildMsg(item)
				b, err := json.Marshal(newMsg)
				if err != nil {
					senderErr = err
					cancel()
					return
				}
				fmt.Printf("\x1b]51;%s\x07", string(b))
			}
		}
	}()

	err := readEach(ctx, os.Stdin, strings.Split(*separators, ","), ch)
	<-done

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if senderErr != nil {
		fmt.Fprintln(os.Stderr, senderErr)
	}
	if err != nil || senderErr != nil {
		os.Exit(1)
	}
}

func makeMsgBuilder(msg []string) func(string) []string {
	if *replaceStr == "" {
		newMsg := make([]string, len(msg)+1)
		copy(newMsg, msg)
		return func(item string) []string {
			newMsg[len(msg)] = item
			return newMsg
		}
	} else {
		template := make([]string, len(msg))
		newMsg := make([]string, len(msg))
		copy(template, msg)
		return func(item string) []string {
			for i := range template {
				newMsg[i] = strings.ReplaceAll(template[i], *replaceStr, item)
			}
			return newMsg
		}
	}
}

// TODO: separators
func readEach(ctx context.Context, r io.Reader, separators []string, dst chan string) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			close(dst)
			return nil
		case dst <- scanner.Text():
		}
	}
	close(dst)
	return scanner.Err()
}
