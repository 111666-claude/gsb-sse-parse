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
// 解析状态逐字节推进，已处理的字节不会被再次扫描，
// 因此结果与分块方式无关，且重复扫描量不随流长度放大。
type Parser struct {
	line    []byte // 当前行尚未遇到行结束符的内容
	heldCR  bool   // 上一段以 \r 结尾，需等下一个字节判断是否为 \r\n
	data    []string
	name    string
	hasName bool
	hasData bool
}

func (p *Parser) resetEvent() {
	p.data = p.data[:0]
	p.name = ""
	p.hasName = false
	p.hasData = false
}

// dispatch 结束当前事件（空行或流结束时调用）。
func (p *Parser) dispatch() (Event, bool) {
	if !p.hasData && !p.hasName {
		return Event{}, false
	}
	name := p.name
	if !p.hasName {
		name = "message"
	}
	ev := Event{Name: name, Data: strings.Join(p.data, "\n")}
	p.resetEvent()
	return ev, true
}

// handleLine 处理一整行（不含行结束符）。
func (p *Parser) handleLine(line string, out *[]Event) {
	if line == "" {
		if ev, ok := p.dispatch(); ok {
			*out = append(*out, ev)
		}
		return
	}
	if line[0] == ':' {
		return // 注释行：不产生也不污染数据
	}
	name, value := line, ""
	if idx := strings.IndexByte(line, ':'); idx >= 0 {
		name, value = line[:idx], line[idx+1:]
		value = strings.TrimPrefix(value, " ")
	}
	switch name {
	case "data":
		p.data = append(p.data, value)
		p.hasData = true
	case "event":
		p.name = value
		p.hasName = true
	}
}

// Feed 吃进一段字节，返回这段字节里凑齐的事件。
func (p *Parser) Feed(chunk string, c *Counter) []Event {
	var out []Event
	start := 0
	if p.heldCR {
		// 上一段末尾的 \r 需要与这段首字节合并判断；只重扫这一个字节。
		c.Scanned++
		p.heldCR = false
		if strings.HasPrefix(chunk, "\n") {
			// 与上一段的 \r 合成 \r\n，消费掉这个 \n。
			start = 1
		}
		p.handleLine(string(p.line), &out)
		p.line = p.line[:0]
	}
	for i := start; i < len(chunk); i++ {
		switch chunk[i] {
		case '\n':
			p.handleLine(string(p.line), &out)
			p.line = p.line[:0]
		case '\r':
			if i+1 < len(chunk) {
				if chunk[i+1] == '\n' {
					i++
				}
				p.handleLine(string(p.line), &out)
				p.line = p.line[:0]
			} else {
				p.heldCR = true
			}
		default:
			p.line = append(p.line, chunk[i])
		}
	}
	return out
}

// Close 处理流结束后还没收尾的内容。
func (p *Parser) Close(c *Counter) []Event {
	_ = c
	var out []Event
	if p.heldCR {
		p.heldCR = false
		p.handleLine(string(p.line), &out)
		p.line = p.line[:0]
	} else if len(p.line) > 0 {
		p.handleLine(string(p.line), &out)
		p.line = p.line[:0]
	}
	if ev, ok := p.dispatch(); ok {
		out = append(out, ev)
	}
	return out
}
