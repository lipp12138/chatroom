# chatroom

给「谁是卧底」直接出词用的 Go 包。默认可以直接出词，也支持从本地 JSON 文件或远程接口加载你自己的词库。

## 安装

```bash
go get github.com/lipp12138/chatroom
```

## 直接出词

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

## 用自己的词库文件

新建 `words.json`：

```json
[
  {"civilian": "苹果", "undercover": "梨"},
  {"civilian": "可乐", "undercover": "雪碧"},
  {"civilian": "火锅", "undercover": "麻辣烫"}
]
```

启动时加载一次：

```go
picker, err := chatroom.LoadFile("words.json")
if err != nil {
	panic(err)
}

pair, err := picker.Pick()
if err != nil {
	panic(err)
}
```

## 用远程接口词库

接口返回同样的 JSON 数组即可：

```go
picker, err := chatroom.LoadURL(context.Background(), "https://example.com/words.json")
if err != nil {
	panic(err)
}

pair, err := picker.Pick()
if err != nil {
	panic(err)
}
```

建议服务启动时加载一次词库并复用 `picker`，不要每次请求都重新读取文件或远程接口。

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
- `chatroom.LoadFile(path)`：从本地 JSON 文件加载词库。
- `chatroom.LoadURL(ctx, url)`：从远程 JSON 接口加载词库。
- `chatroom.New(pairs)`：从代码里的 `[]chatroom.Pair` 创建词库。
- `picker.Pick()`：从自定义词库随机返回一组词。
- `picker.Round(playerCount, undercoverCount)`：从自定义词库生成一局玩家分词。
- `chatroom.All()`：返回内置词库副本。
