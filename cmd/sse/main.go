// Command sse 从标准输入读 SSE 流并按行打印事件。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/sse-parse/sse"
)

func emit(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func collect(events []sse.Event) []map[string]string {
	var out []map[string]string
	for _, ev := range events {
		out = append(out, map[string]string{"event": ev.Name, "data": ev.Data})
	}
	return out
}

func run(sample string) {
	p := &sse.Parser{}
	c := &sse.Counter{}
	switch sample {
	case "split":
		first := p.Feed("data: a", c)
		second := p.Feed("\n\ndata: b\n\n", c)
		emit(map[string]any{"events": collect(append(first, second...))})
	case "multiline":
		emit(map[string]any{"events": collect(p.Feed("data: a\ndata: b\n\n", c))})
	case "crlf":
		emit(map[string]any{"events": collect(p.Feed("data: x\r\n\r\n", c))})
	case "comment":
		emit(map[string]any{"events": collect(p.Feed("data: x\n: keep-alive\n\n", c))})
	case "flush":
		p.Feed("data: tail", c)
		emit(map[string]any{"events": collect(p.Close(c))})
	case "work":
		var sb strings.Builder
		for i := 0; i < 2000; i++ {
			sb.WriteString(fmt.Sprintf("data: e%d\n\n", i))
		}
		raw := sb.String()
		n := 0
		for i := 0; i < len(raw); i += 5 {
			end := i + 5
			if end > len(raw) {
				end = len(raw)
			}
			n += len(p.Feed(raw[i:end], c))
		}
		emit(map[string]int{"events": n, "scanned": c.Scanned})
	default:
		fmt.Fprintln(os.Stderr, "未知场景："+sample)
		os.Exit(2)
	}
}

func main() {
	sample := flag.String("sample", "", "split|multiline|crlf|comment|flush|work")
	flag.Parse()

	if *sample != "" {
		run(*sample)
		return
	}
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	p := &sse.Parser{}
	c := &sse.Counter{}
	for _, ev := range append(p.Feed(string(raw), c), p.Close(c)...) {
		emit(ev)
	}
}
