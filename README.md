# chatroom

「谁是卧底」出词 Go 包。

运行时只有两个动作：

```go
chatroom.Update("/data/ciku/a.txt") // 更新词库时调用
pair, err := chatroom.Pick()        // 出词时调用，不读文件
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
	if err := chatroom.Update("/data/ciku/a.txt"); err != nil {
		log.Println("词库更新失败，已使用默认词库：", err)
	}

	pair, err := chatroom.Pick()
	if err != nil {
		panic(err)
	}

	fmt.Println("平民词：", pair.Civilian)
	fmt.Println("卧底词：", pair.Undercover)
}
```

## TXT 词库格式

`/data/ciku/a.txt`：

```txt
苹果,梨
可乐,雪碧
火锅,麻辣烫
```

也支持这些分隔符：英文逗号 `,`、中文逗号 `，`、竖线 `|`、Tab。

空行和 `#` 开头的注释行会被忽略。

## JSON 词库格式

如果文件后缀是 `.json`，也可以用 JSON：

```json
[
  {"civilian": "苹果", "undercover": "梨"},
  {"civilian": "可乐", "undercover": "雪碧"},
  {"civilian": "火锅", "undercover": "麻辣烫"}
]
```

## 行为

- `chatroom.Pick()` 只从内存出词，不读取文件。
- `chatroom.Update(path)` 才会读取本地词库文件。
- `Update` 成功后，后续 `Pick` 使用新词库。
- `Update` 失败时，自动回退默认词库，并返回 `err` 方便打日志。
- 不停服务更新词库时，改完文件后再调用一次 `Update(path)` 即可。

推荐更新文件时先写临时文件，再重命名覆盖正式文件，避免服务读到写了一半的内容。
