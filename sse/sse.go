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
	buf     string   // 还没凑成一整行的字节
	scan    int      // buf 里已确认没有行结束符的前缀长度
	data    []string // 当前事件已收到的 data 行，按出现顺序
	name    string   // 当前事件名，空表示未设置
	hasData bool     // 当前事件是否收到过 data 行
}

func (p *Parser) dispatch() Event {
	name := p.name
	if name == "" {
		name = "message"
	}
	return Event{Name: name, Data: strings.Join(p.data, "\n")}
}

func (p *Parser) resetEvent() {
	p.data = nil
	p.name = ""
	p.hasData = false
}

// line 消费一整行；空行结束当前事件并派发。
func (p *Parser) line(line string, out []Event) []Event {
	if line == "" {
		if p.hasData {
			out = append(out, p.dispatch())
		}
		p.resetEvent()
		return out
	}
	if line[0] == ':' {
		return out // 注释行，不产生也不污染数据
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
	}
	return out
}

// drain 从 buf 里切出完整的行逐行消费。final 为假时，末尾的 \r 可能是
// 被切块拆开的 \r\n，留下次再判；final 为真时直接算行结束。
func (p *Parser) drain(final bool) []Event {
	var out []Event
	i, lineStart := p.scan, 0
hold:
	for i < len(p.buf) {
		switch p.buf[i] {
		case '\n':
			out = p.line(p.buf[lineStart:i], out)
			i++
			lineStart = i
		case '\r':
			if i+1 == len(p.buf) && !final {
				break hold
			}
			out = p.line(p.buf[lineStart:i], out)
			i++
			if i < len(p.buf) && p.buf[i] == '\n' {
				i++
			}
			lineStart = i
		default:
			i++
		}
	}
	p.buf = p.buf[lineStart:]
	p.scan = i - lineStart
	return out
}

// Feed 吃进一段字节，返回这段字节里凑齐的事件。
func (p *Parser) Feed(chunk string, c *Counter) []Event {
	c.Scanned += len(p.buf) - p.scan // 至多末尾待定的 \r 会被重新检查
	p.buf += chunk
	return p.drain(false)
}

// Close 处理流结束后还没收尾的内容。
func (p *Parser) Close(c *Counter) []Event {
	c.Scanned += len(p.buf) - p.scan
	out := p.drain(true)
	if len(p.buf) > 0 {
		out = p.line(p.buf, out) // 末尾没有行结束符的最后一行
		p.buf = ""
		p.scan = 0
	}
	if p.hasData {
		out = append(out, p.dispatch())
	}
	p.resetEvent()
	return out
}
