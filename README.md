# chatroom

「谁是卧底」出词 Go 包。

整个包只保留一个出词方法：

```go
pair, err := chatroom.Pick()
pair, err := chatroom.Pick("/data/ciku/a.json")
```

## 安装

```bash
go get github.com/lipp12138/chatroom
```

## 使用

```go
package main

import (
	"fmt"
	"log"

	"github.com/lipp12138/chatroom"
)

func main() {
	pair, err := chatroom.Pick("/data/ciku/a.json")
	if err != nil {
		log.Println("词库提示：", err)
	}

	fmt.Println("平民词：", pair.Civilian)
	fmt.Println("卧底词：", pair.Undercover)
}
```

## 词库格式

`/data/ciku/a.json`：

```json
[
  {"civilian": "苹果", "undercover": "梨"},
  {"civilian": "可乐", "undercover": "雪碧"},
  {"civilian": "火锅", "undercover": "麻辣烫"}
]
```

## 行为

- 不传文件：使用内置默认词库。
- 传文件且文件正常：使用本地词库。
- 文件不存在：使用默认词库，`err == nil`。
- 文件读取失败、JSON 写错、没有有效词：使用默认词库，同时返回 `err`，方便你打日志。
- 文件会缓存；文件内容修改后，下次 `Pick` 会自动重新加载。

不停服务更新词库时，建议先写临时文件，再重命名覆盖正式文件，避免服务读到写了一半的 JSON。
