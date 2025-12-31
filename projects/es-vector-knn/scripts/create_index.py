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
    parser.add_argument(
        "--mapping",
        default="data/index.json",
        help="Mapping JSON path (default: data/index.json)",
    )
    args = parser.parse_args()

    es_url = args.es_url or (getenv("ES_URL") or "http://localhost:9200")
    index = args.index or (getenv("INDEX") or "items")

    mapping = read_json(args.mapping)

    # Best-effort delete.
    _ = http_request(
        "DELETE", f"{es_url}/{index}", None, content_type=None, ignore_statuses={404}
    )

    resp = http_request("PUT", f"{es_url}/{index}", json.dumps(mapping).encode("utf-8"))
    print(resp)
    return 0


def read_json(path: str) -> object:
    try:
        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)
    except Exception as e:
        print(f"failed to read json: {path}: {e}", file=sys.stderr)
        raise


def http_request(
    method: str,
    url: str,
    body: bytes | None,
    content_type: str | None = "application/json",
    ignore_statuses: set[int] | None = None,
) -> str:
    headers = {}
    if content_type is not None:
        headers["Content-Type"] = content_type
    req = urllib.request.Request(url, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return resp.read().decode("utf-8")
    except urllib.error.HTTPError as e:
        if ignore_statuses and e.code in ignore_statuses:
            return ""
        msg = e.read().decode("utf-8", errors="replace")
        print(
            f"http error: {method} {url}: status={e.code} body={msg}", file=sys.stderr
        )
        raise
    except Exception as e:
        print(f"request failed: {method} {url}: {e}", file=sys.stderr)
        raise


def getenv(key: str) -> str | None:
    import os

    return os.environ.get(key)


if __name__ == "__main__":
    raise SystemExit(main())
