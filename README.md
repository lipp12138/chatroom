# chatroom

「谁是卧底」出词 Go 包。

运行时只有两个动作：

```go
chatroom.Update("/data/ciku/a.txt") // 更新词库时调用
pair, err := chatroom.Pick("room-1") // 出词时调用，不读文件
```

## 安装

```bash
go get github.com/lipp12138/chatroom
```

## 已备好词库

仓库里已经准备了一份可直接使用的 txt 题库：

```text
data/ciku/a.txt
```

里面包含 500+ 组题，覆盖食物饮品、日常用品、社交名场面、职场沟通、网络生活、悬疑推理、影视综艺感等场景。你可以直接加载：

```go
_ = chatroom.Update("data/ciku/a.txt")
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

	pair, err := chatroom.Pick("room-1001")
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
- `chatroom.Pick("房间号")` 会按房间记录最近出过的词；同一个房间最近 5 次不会重复。
- 不同房间的出词记录互不影响。
- 房间记录只存在内存里，超过 2 小时没有使用会自动清理。
- `chatroom.Update(path)` 才会读取本地词库文件。
- `Update` 成功后，后续 `Pick` 使用新词库。
- `Update` 失败时，自动回退默认词库，并返回 `err` 方便打日志。
- 不停服务更新词库时，改完文件后再调用一次 `Update(path)` 即可。

推荐更新文件时先写临时文件，再重命名覆盖正式文件，避免服务读到写了一半的内容。
