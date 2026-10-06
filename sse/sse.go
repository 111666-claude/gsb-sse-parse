// Package sse 增量解析 Server-Sent Events 流。
package sse

import "strings"

// Event 是一条解析出来的事件。
type Event struct {
	Name string `json:"event"`
	Data string `json:"data"`
}

// Counter 统计重复扫描过的字节数。
type Counter struct {
	Scanned int
}

// Parser 累积字节流并吐出完整事件。
type Parser struct {
	buf string
}

func field(line, name string) (string, bool) {
	prefix := name + ":"
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	value := strings.TrimPrefix(line, prefix)
	value = strings.TrimPrefix(value, " ")
	return value, true
}

func parseBlock(block string) (Event, bool) {
	name := "message"
	data := ""
	has := false
	for _, line := range strings.Split(block, "\n") {
		if value, ok := field(line, "data"); ok {
			data = value
			has = true
			continue
		}
		if value, ok := field(line, "event"); ok {
			name = value
			continue
		}
		if strings.HasPrefix(line, ":") {
			data = line
			has = true
		}
	}
	if !has {
		return Event{}, false
	}
	return Event{Name: name, Data: data}, true
}

// Feed 吃进一段字节，返回这段字节里凑齐的事件。
func (p *Parser) Feed(chunk string, c *Counter) []Event {
	c.Scanned += len(p.buf)
	p.buf += chunk
	var out []Event
	for {
		idx := strings.Index(chunk, "\n\n")
		if idx < 0 {
			break
		}
		block := p.buf[:idx]
		p.buf = p.buf[idx+2:]
		chunk = chunk[idx+2:]
		if ev, ok := parseBlock(block); ok {
			out = append(out, ev)
		}
	}
	return out
}

// Close 处理流结束后还没收尾的内容。
func (p *Parser) Close(c *Counter) []Event {
	_ = c
	return nil
}
