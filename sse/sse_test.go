package sse

import "testing"

func TestSingleDataLine(t *testing.T) {
	p := &Parser{}
	out := p.Feed("data: hello\n\n", &Counter{})
	if len(out) != 1 || out[0].Data != "hello" {
		t.Fatalf("应解析出一条 data=hello，得到 %+v", out)
	}
}

func TestDefaultEventName(t *testing.T) {
	p := &Parser{}
	out := p.Feed("data: x\n\n", &Counter{})
	if len(out) != 1 || out[0].Name != "message" {
		t.Fatalf("默认事件名应是 message，得到 %+v", out)
	}
}

func TestCustomEventName(t *testing.T) {
	p := &Parser{}
	out := p.Feed("event: ping\ndata: x\n\n", &Counter{})
	if len(out) != 1 || out[0].Name != "ping" {
		t.Fatalf("事件名应是 ping，得到 %+v", out)
	}
}

func TestTwoBlocksGiveTwoEvents(t *testing.T) {
	p := &Parser{}
	out := p.Feed("data: a\n\ndata: b\n\n", &Counter{})
	if len(out) != 2 {
		t.Fatalf("两个块应给两条事件，得到 %+v", out)
	}
}

func TestEmptyStream(t *testing.T) {
	p := &Parser{}
	if out := p.Feed("", &Counter{}); len(out) != 0 {
		t.Fatalf("空流不该有事件，得到 %+v", out)
	}
}
