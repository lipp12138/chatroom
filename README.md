# chatroom

`chatroom` 是一个给「谁是卧底」出词用的 Go 包。它内置中文相近词对，初始化时建立索引，之后可以快速随机抽词或直接生成一局玩家词牌。

## 安装

```bash
go get github.com/lipp12138/chatroom
```

## 快速使用

```go
package main

import (
	"fmt"

	"github.com/lipp12138/chatroom"
)

func main() {
	bank := keywords.Default()

	pair, _ := bank.Pick(
		keywords.WithCategory("food"),
		keywords.WithDifficulty(keywords.Easy),
	)
	fmt.Printf("平民词：%s，卧底词：%s\n", pair.Civilian, pair.Undercover)

	round, _ := bank.Round(8, 2)
	for _, role := range round.Roles {
		fmt.Printf("%d 号：%s\n", role.Player, role.Word)
	}
}
```

## API

- `Default()`：返回内置词库。
- `New(pairs []Pair)`：使用自定义词库。
- `LoadJSON(r io.Reader)`：从 JSON 数组加载词库。
- `Pick(opts ...Option)`：随机抽取词对。
- `Round(playerCount, undercoverCount int, opts ...Option)`：生成一局玩家分词。
- `WithCategory(category string)`：按分类过滤。
- `WithDifficulty(difficulty Difficulty)`：按难度过滤。

内置分类包括：`food`、`daily`、`travel`、`sports`、`entertainment`、`people`、`nature`、`place`、`internet`、`study`、`art`、`life`、`festival`。

## 性能

词库在 `New` 时完成校验、去重和索引预计算。初始化后，普通抽词和带分类/难度筛选的抽词都只需要在候选索引中随机命中，适合在游戏服务里高频调用。
