import os
from dataclasses import asdict, dataclass
from pathlib import Path

import databento
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
    timestamp: str  # raw dataset timestamp string (e.g., ISO-8601)
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


def unpack_dbn_from_file(dbn_path: str) -> None:
    df = databento.DBNStore.from_file(path=dbn_path).to_df()
    for index, row in df.iterrows():
        print(f"[{index}]: {row}")
        # symbol: str = row.symbol[:2]

        # if "-" in row.symbol or symbol not in self.symbols:
        #     continue

        # if index not in data[symbol] or row.volume > data[symbol][index].volume:
        #     data[symbol][index] = Candle(
        #         symbol, self.timeframe, index.to_pydatetime(), row.open, row.high, row.low, row.close, row.volume
        #     )

def request_data():
    dataset = "GLBX.MDP3"
    product = "ES"
    start = "2025-11-10"
    end = "2025-12-10"

# Create a historical client
    client = databento.Historical(os.getenv("DB_API_KEY"))

# Request OHLCV-1d data for the continuous contract
    data = client.timeseries.get_range(
        dataset=dataset,
        schema="ohlcv-1d",
        symbols=f"{product}.v.0",
        stype_in="continuous",
        start=start,
        end=end,
    )

# Convert to DataFrame
    df = data.to_df()

    print(df)


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
    # unpack_dbn_from_file("glbx-mdp3-20240829-20240831.ohlcv-1m.dbn.zst")
    request_data()


if __name__ == "__main__":
    main()
