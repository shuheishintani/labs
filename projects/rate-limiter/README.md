# rate-limiter

## 概要

Go で簡易な rate limiter（Token Bucket / Fixed Window）を実装し、アルゴリズムの違い（バースト許容、境界問題、精度など）を理解する。

## 目的

- Token Bucket / Fixed Window を in-memory のライブラリとして実装し、挙動の違いを観察できる状態にする
- `Allow(key)` の API を通して、許可/拒否と待ち時間の扱いを理解する

## 対象外

- 分散環境での整合性（Redis 等を使った共有状態）
- HTTP サーバとしての提供（この時点ではライブラリ + デモに留める）
- 永続化・運用（監視、レート調整 UI など）

## 使い方

### デモ

指定したアルゴリズムを一定間隔で連続実行し、`allowed/denied` と `retryAfter` を表示します。

実行：

```bash
go run ./cmd/server -algo tokenbucket
```

フラグ例：

```bash
# Token Bucket: 許可/拒否の切り替わりを観察
go run ./cmd/server -algo tokenbucket -rate 2 -burst 2 -interval 100ms -count 30

# Fixed Window: 窓内の上限到達と retryAfter を観察
go run ./cmd/server -algo fixedwindow -limit 2 -window 1s -interval 100ms -count 10

# Token Bucket: 拒否時に待って再試行する挙動を観察
go run ./cmd/server -algo tokenbucket -rate 1 -burst 1 -interval 50ms -count 50 -sleep-on-deny
```

## 学び・気づき

- **Token Bucket**: 平均レート（`rate`）に加えて、最大容量（`burst`）でバーストを自然に許容できる
- **Fixed Window**: 実装は単純だが、窓境界にリクエストが寄ると「短時間に多く通る」挙動が起きやすい
- 拒否時に返す `retryAfter` は「次に通せるまでの目安時間」
  - Token Bucket: 次の 1 トークンが貯まるまでの見積もり
  - Fixed Window: 現在の窓が終わるまでの見積もり
- `-sleep-on-deny` を付けると、クライアントが待って再試行する挙動を疑似的に観察できる
