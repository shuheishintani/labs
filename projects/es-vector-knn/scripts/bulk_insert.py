#!/usr/bin/env python3
import argparse
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
    parser.add_argument(
        "--ndjson",
        default="data/bulk.ndjson",
        help="Bulk NDJSON path (default: data/bulk.ndjson)",
    )
    args = parser.parse_args()

    es_url = args.es_url or (getenv("ES_URL") or "http://localhost:9200")
    index = args.index or (getenv("INDEX") or "items")

    try:
        with open(args.ndjson, "rb") as f:
            payload = f.read()
    except Exception as e:
        print(f"failed to read ndjson: {args.ndjson}: {e}", file=sys.stderr)
        return 1

    url = f"{es_url}/{index}/_bulk?refresh=true"
    req = urllib.request.Request(
        url,
        data=payload,
        method="POST",
        headers={"Content-Type": "application/x-ndjson"},
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            body = resp.read().decode("utf-8")
            print(body)
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
