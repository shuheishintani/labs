#!/usr/bin/env python3
import argparse
import json
import sys
import urllib.error
import urllib.request


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--es-url",
        default=None,
        help="Elasticsearch URL (default: env ES_URL or http://localhost:9200)",
    )
    parser.add_argument(
        "--index", default=None, help="Index name (default: env INDEX or items)"
    )
    parser.add_argument("-k", type=int, default=3, help="k (default: 3)")
    parser.add_argument(
        "--num-candidates", type=int, default=20, help="num_candidates (default: 20)"
    )
    parser.add_argument("--category", default="", help="optional category filter")
    args = parser.parse_args()

    es_url = args.es_url or (getenv("ES_URL") or "http://localhost:9200")
    index = args.index or (getenv("INDEX") or "items")

    # query_vector は手書き。docs の embedding と「近い」方向のベクトルを入れてある。
    query_vector = [0.9, 0.1, 0.0]

    knn = {
        "field": "embedding",
        "query_vector": query_vector,
        "k": args.k,
        "num_candidates": args.num_candidates,
    }
    if args.category:
        knn["filter"] = [{"term": {"category": args.category}}]

    body = {
        "knn": knn,
        "_source": ["title", "category"],
    }

    url = f"{es_url}/{index}/_search"
    payload = json.dumps(body).encode("utf-8")

    req = urllib.request.Request(
        url, data=payload, method="POST", headers={"Content-Type": "application/json"}
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            print(resp.read().decode("utf-8"))
            return 0
    except urllib.error.HTTPError as e:
        msg = e.read().decode("utf-8", errors="replace")
        print(f"http error: POST {url}: status={e.code} body={msg}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"request failed: POST {url}: {e}", file=sys.stderr)
        return 1


def getenv(key: str) -> str | None:
    import os

    return os.environ.get(key)


if __name__ == "__main__":
    raise SystemExit(main())
