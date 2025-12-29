# consistent-hash

## 概要

Go で consistent hashing（virtual nodes あり）を実装し、ノード追加/削除時にキーの割り当てが大きく崩れない性質を確認する。

## 目的

- consistent hash ring の最小実装（Add/Remove/Get）を作る
- virtual nodes（replicas）が分布に効く理由を理解する

## 対象外

- 重み付きノード（weight）
- レプリケーション（複数ノードを返す）
- 分散環境でのメンバーシップ管理（障害検知、合意など）

## 使い方

実行:

```bash
go run ./cmd/server
```

フラグ例:

```bash
go run ./cmd/server -replicas 200 -nodes node-a,node-b,node-c -key user:123
```

テスト:

```bash
go test ./...
```

API 例:

```go
ch, _ := consistenthash.New(100) // replicas
ch.Add("node-a", "node-b", "node-c")

node, ok := ch.Get("user:123")
_ = node
_ = ok
```

## 学び・気づき

- ring は「ソート済みの hash の配列」と「hash→node の対応」で表現でき、`sort.Search` で lookup できる
- `replicas` を増やすと（リング上の点が増えるため）割り当てが滑らかになりやすい
- Get は「空なら ok=false」を返すようにし、呼び出し側で扱いやすくする
