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
	"unicode/utf8"
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

	sepRunes, err := convertSeparators(*separators)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = readEach(ctx, os.Stdin, sepRunes, ch)
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

func convertSeparators(separators string) ([]rune, error) {
	runes := make([]rune, 0, 5)
	for _, s := range strings.Split(separators, ",") {
		switch s {
		case "space":
			runes = append(runes, ' ')
		case "tab":
			runes = append(runes, '\t')
		case "newline":
			runes = append(runes, '\r', '\n')
		case "nul":
			runes = append(runes, '\x00')
		default:
			return nil, fmt.Errorf("unknown separator '%s'", s)
		}
	}
	return runes, nil
}

func readEach(ctx context.Context, r io.Reader, separators []rune, dst chan string) error {
	scanner := bufio.NewScanner(r)
	isSep := func(r rune) bool {
		for _, s := range separators {
			if s == r {
				return true
			}
		}
		return false
	}
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return scanWords(isSep, data, atEOF)
	})
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

// https://golang.org/src/bufio/scan.go?s=13096:13174#L380
func scanWords(isSep func(rune) bool, data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip separators.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !isSep(r) {
			break
		}
	}
	// Scan until non-separator character, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if isSep(r) {
			return i + width, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}
