# chatroom

给「谁是卧底」直接出词用的 Go 包。只保留一个出词方法：不传文件就用默认词库，传本地 JSON 文件就优先用这个词库，文件不存在时自动回退默认词库。

## 安装

```bash
go get github.com/lipp12138/chatroom
```

## 使用默认词库

```go
package main

import (
	"fmt"

	"github.com/lipp12138/chatroom"
)

func main() {
	pair, err := chatroom.Pick()
	if err != nil {
		panic(err)
	}

	fmt.Println("平民词：", pair.Civilian)
	fmt.Println("卧底词：", pair.Undercover)
}
```

## 使用本地词库

```go
pair, err := chatroom.Pick("/data/ciku/a.json")
if err != nil {
	panic(err)
}
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
- `chatroom.Pick("/data/ciku/a.json")`：文件存在就使用这个词库。
- 文件不存在：自动使用默认词库。
- 文件存在但 JSON 写错：返回错误。
- 文件会自动缓存；文件内容修改后，下次出词会重新加载。
