# chatroom

给「谁是卧底」直接出词用的 Go 包。默认可以直接出词，也可以通过配置读取本地 JSON 词库；配置文件不存在时自动使用内置默认词库。

## 安装

```bash
go get github.com/lipp12138/chatroom
```

## 最简单用法

```go
package main

import (
	"fmt"

	"github.com/lipp12138/chatroom"
)

func main() {
	pair := chatroom.Pick()
	fmt.Println("平民词：", pair.Civilian)
	fmt.Println("卧底词：", pair.Undercover)
}
```

## 按配置读取词库

比如你想优先读取 `/data/ciku/a.json`，没有这个文件就用默认词库：

```go
picker, err := chatroom.Load(chatroom.Config{
	File: "/data/ciku/a.json",
})
if err != nil {
	panic(err)
}

pair, err := picker.Pick()
if err != nil {
	panic(err)
}
```

建议服务启动时加载一次 `picker`，后面一直复用，不要每次请求都重新读文件。

## 词库文件格式

`/data/ciku/a.json` 内容示例：

```json
[
  {"civilian": "苹果", "undercover": "梨"},
  {"civilian": "可乐", "undercover": "雪碧"},
  {"civilian": "火锅", "undercover": "麻辣烫"}
]
```

## 直接读取文件

如果你不需要“文件不存在就默认”的逻辑，也可以直接读取文件：

```go
picker, err := chatroom.LoadFile("/data/ciku/a.json")
```

## 远程接口词库

接口返回同样的 JSON 数组即可：

```go
picker, err := chatroom.LoadURL(context.Background(), "https://example.com/words.json")
```

## 生成一局分词

```go
round, err := picker.Round(8, 2)
if err != nil {
	panic(err)
}

for _, role := range round.Roles {
	fmt.Printf("%d号：%s\n", role.Player, role.Word)
}
```

## API

- `chatroom.Pick()`：从内置词库随机返回一组词。
- `chatroom.Load(config)`：按配置加载词库；文件为空或不存在时使用默认词库。
- `chatroom.LoadFile(path)`：从本地 JSON 文件加载词库；文件不存在会返回错误。
- `chatroom.LoadURL(ctx, url)`：从远程 JSON 接口加载词库。
- `chatroom.New(pairs)`：从代码里的 `[]chatroom.Pair` 创建词库。
- `picker.Pick()`：从已加载词库随机返回一组词。
- `picker.Round(playerCount, undercoverCount)`：从已加载词库生成一局玩家分词。
- `chatroom.All()`：返回内置词库副本。
