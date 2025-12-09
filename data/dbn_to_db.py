import os
from pathlib import Path

import psycopg


def list_dbn_files(directory: str = ".") -> list[str]:
    base = Path(directory)
    files: list[str] = []
    for entry in base.iterdir():
        if entry.is_file() and entry.name.endswith(".dbn.zst"):
            files.append(str(entry))
    return sorted(files)


def print_candles(db_url: str) -> None:
    with psycopg.connect(db_url) as conn:
        with conn.cursor() as cur:
            cur.execute("SELECT * FROM candles")
            for row in cur.fetchall():
                print(row)


def main() -> None:
    files = list_dbn_files(".")
    if files:
        print("Found .dbn.zst files:")
        for path in files:
            print(f" - {path}")
    else:
        print("No .dbn.zst files found in current directory")

    db_url = os.getenv("LOCAL_DB_URL")
    if not db_url:
        raise RuntimeError("DB_URL environment variable is not set")

    print("Querying candles table...")
    print_candles(db_url)


if __name__ == "__main__":
    main()
