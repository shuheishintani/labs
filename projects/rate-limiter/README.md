# rate-limiter

## 概要

Go で簡易な rate limiter を実装し、アルゴリズムの違い（バースト許容、境界問題、精度など）を理解する。

## 目的

- Token Bucket を in-memory のライブラリとして実装し、挙動を観察できる状態にする
- `Allow(key)` の API を通して、許可/拒否と待ち時間の扱いを理解する

## 対象外

- 分散環境での整合性（Redis 等を使った共有状態）
- HTTP サーバとしての提供（この時点ではライブラリ + デモに留める）
- 永続化・運用（監視、レート調整 UI など）

## 使い方

### デモ

Token Bucket を一定間隔で連続実行し、`allowed/denied` と `retryAfter` を表示します。

```bash
go run ./cmd/demo
```

オプション例：

```bash
# 例1: rate=2, burst=2 で 100ms 間隔・30回実行（許可/拒否の切り替わりを観察）
go run ./cmd/demo -rate 2 -burst 2 -interval 100ms -count 30

# 例2: rate=1, burst=1 で 50ms 間隔・50回実行（拒否時に待って再試行する挙動を観察）
go run ./cmd/demo -rate 1 -burst 1 -interval 50ms -count 50 -sleep-on-deny
```

## 学び・気づき

### rate と burst の違い（Token Bucket）

- rate: 回復速度（長時間平均で毎秒どれだけ通せるか）
- burst: 貯められる上限（瞬間的にどれだけ連続で通せるか）
