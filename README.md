# chatroom

给「谁是卧底」直接出词用的 Go 包。只保留一个出词方法：不传文件就用默认词库，传本地 JSON 文件就优先用这个词库。文件不存在、读取失败或 JSON 写错时，会自动回退默认词库，并通过 `err` 返回提示。

## 安装

```bash
go get github.com/lipp12138/chatroom
```

## 使用默认词库

```go
pair, err := chatroom.Pick()
if err != nil {
	log.Println("词库提示：", err)
}

fmt.Println("平民词：", pair.Civilian)
fmt.Println("卧底词：", pair.Undercover)
```

## 使用本地词库

```go
pair, err := chatroom.Pick("/data/ciku/a.json")
if err != nil {
	log.Println("词库提示：", err)
}

fmt.Println("平民词：", pair.Civilian)
fmt.Println("卧底词：", pair.Undercover)
```

`/data/ciku/a.json` 示例：

```json
[
  {"civilian": "苹果", "undercover": "梨"},
  {"civilian": "可乐", "undercover": "雪碧"},
  {"civilian": "火锅", "undercover": "麻辣烫"}
]
```

说明：

- `chatroom.Pick()`：使用内置默认词库。
- `chatroom.Pick("/data/ciku/a.json")`：文件存在且格式正确时使用这个词库。
- 文件不存在：自动使用默认词库，`err` 为 `nil`。
- 文件读取失败、JSON 写错或没有有效词：自动使用默认词库，同时返回 `err`，方便你打日志提示。
- 文件会自动缓存；文件内容修改后，下次出词会重新加载。
