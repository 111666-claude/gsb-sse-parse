# sse-parse

Server-Sent Events 流解析：从标准输入增量吃字节，一条事件打一行 JSON。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/sse --sample=split
go run ./cmd/sse --sample=multiline
go run ./cmd/sse --sample=crlf
go run ./cmd/sse --sample=comment
go run ./cmd/sse --sample=flush
go run ./cmd/sse --sample=work
echo -n $'data: hi\n\n' | go run ./cmd/sse
```

## 口径

- **分块无关**：同一段字节流，无论一次喂进来还是切成任意多块喂进来，吐出的事件序列必须逐条一致。
- **事件分隔**：空行结束一条事件；`\n`、`\r\n`、`\r` 都算行结束。
- **多行数据**：同一个事件里多条 `data:` 用换行按出现顺序拼起来。
- **事件名**：`event:` 给出的名字；没给就是 `message`。
- **注释**：以 `:` 开头的行只是注释，不产生也不污染数据。
- **收尾**：流结束时若还有没被空行结束的事件，也要吐出来。

## 不变量

- 一条事件最多一个事件名，数据按出现顺序拼接。
- 解析结果与分块方式无关。
- `scanned`（重复扫描的字节数）不随流长度放大：两千条事件时不超过 9000。

## 输出契约

每条事件打印一行 JSON：`event`、`data`。
`--sample=work` 只打印 `{"events": N, "scanned": N}`。
