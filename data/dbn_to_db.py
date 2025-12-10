import os
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

import psycopg
from psycopg.rows import dict_row


@dataclass
class Candle:
    market: str
    symbol: str
    timeframe: str
    open: float
    high: float
    low: float
    close: float
    volume: int
    timestamp: Any
    id: int | None = None


def list_dbn_files(directory: str = ".") -> list[str]:
    base = Path(directory)
    files: list[str] = []
    for entry in base.iterdir():
        if entry.is_file() and entry.name.endswith(".dbn.zst"):
            files.append(str(entry))
    return sorted(files)


def print_candles(db_url: str) -> None:
    with psycopg.connect(db_url) as conn, conn.cursor() as cur:
        cur.execute("SELECT * FROM candles")
        for row in cur.fetchall():
            print(row)


def insert_candle(db_url: str, candle: Candle) -> int:
    payload = asdict(candle)
    payload.pop("id", None)

    sql = """
        INSERT INTO candles (
            market, symbol, timeframe,
            open, high, low, close,
            volume, timestamp
        ) VALUES (
            %(market)s, %(symbol)s, %(timeframe)s,
            %(open)s, %(high)s, %(low)s, %(close)s,
            %(volume)s, %(timestamp)s
        ) RETURNING id
    """

    with psycopg.connect(db_url, row_factory=dict_row) as conn, conn.cursor() as cur:
        cur.execute(sql, payload)
        row = cur.fetchone()
        return int(row["id"])


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
