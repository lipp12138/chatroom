# chatroom

给「谁是卧底」直接出词用的 Go 包。没有分类、没有难度、没有复杂配置，引入后直接拿一组词。

## 安装

```bash
go get github.com/lipp12138/chatroom
```

## 使用

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

如果你要直接生成一局玩家分词：

```go
round, err := chatroom.Round(8, 2)
if err != nil {
	panic(err)
}

for _, role := range round.Roles {
	fmt.Printf("%d号：%s\n", role.Player, role.Word)
}
```

## API

- `chatroom.Pick()`：随机返回一组词。
- `chatroom.Round(playerCount, undercoverCount)`：生成一局玩家分词。
- `chatroom.All()`：返回内置词库副本。
