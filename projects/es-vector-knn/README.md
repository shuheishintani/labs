# es-vector-knn

## 概要

Elasticsearch の `dense_vector` と `knn` クエリを使って、ベクトル近傍探索（kNN/ANN）の基本を確認する。

## 目的

- `dense_vector` + `index: true` のインデックスを作る
- `knn` クエリで近傍検索できることを確認する
- `k` / `num_candidates` / `filter` の意味を体感する

## 対象外

- embedding の生成（今回は手書きベクトル）
- 本格的な性能チューニング、多ノード構成

## 使い方

起動:

```bash
docker compose up -d
```

インデックス作成とデータ投入:

```bash
python3 scripts/create_index.py
python3 scripts/bulk_insert.py
```

kNN 検索:

```bash
python3 scripts/knn_search.py
python3 scripts/knn_search.py -k 3 --num-candidates 50
python3 scripts/knn_search.py -k 3 --num-candidates 50 --category tech
```

停止:

```bash
docker compose down -v
```

## 要点

- `k`: 返す近傍の件数
- `num_candidates`: ANN で「候補として探索する件数」（増やすほど精度が上がりやすいが遅くなりやすい）
- `filter`: 近傍探索の対象集合を絞る（例: カテゴリで絞ってから近いものを探す）
