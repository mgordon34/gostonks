import operator
import os
import time
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
    i = None
    a = None
    for index, row in df.iterrows():
        i = index
        a = row
        # print(f"[{index}]: {row}")
        # symbol: str = row.symbol[:2]

        # if "-" in row.symbol or symbol not in self.symbols:
        #     continue

        # if index not in data[symbol] or row.volume > data[symbol][index].volume:
        #     data[symbol][index] = Candle(
        #         symbol, self.timeframe, index.to_pydatetime(), row.open, row.high, row.low, row.close, row.volume
        #     )

    print(f"[candle]: {a}")
    c = Candle(
        market="futures",
        symbol=row.symbol.split(".")[0],
        timeframe="1m",
        open=row.open,
        high=row.high,
        low=row.low,
        close=row.close,
        volume=row.volume,
        timestamp=index.to_pydatetime(),
    )
    insert_candle(os.getenv("LOCAL_DB_URL"), c)

def request_data(symbol: str, start_date: str, end_date: str):
    dataset = "GLBX.MDP3"

    client = databento.Historical(os.getenv("DB_API_KEY"))

    data = client.timeseries.get_range(
        dataset=dataset,
        schema="ohlcv-1m",
        symbols=f"{symbol}.v.0",
        stype_in="continuous",
        start=start_date,
        end=end_date,
    )

    data.to_file(f"{dataset}-{symbol}-{start_date}-{end_date}.dbn.zst")
    df = data.to_df()

    print(df)

def get_dbn_historical_batch(symbol: str, start_date: str, end_date: str):
    dataset = "GLBX.MDP3"

    client = databento.Historical(os.getenv("DB_API_KEY"))
    new_job = client.batch.submit_job(
        dataset=dataset,
        symbols=f"{symbol}.v.0",
        schema="ohlcv-1m",
        stype_in="continuous",
        split_duration="month",
        start=start_date,
        end=end_date,
    )

    new_job_id: str = new_job["id"]

    while True:
        done_jobs = list(map(operator.itemgetter("id"), client.batch.list_jobs("done")))
        if new_job_id in done_jobs:
            break  # Exit the loop to continue
        time.sleep(1.0)

    client.batch.download(
        job_id=new_job_id,
        output_dir=Path.cwd(),
    )


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
    unpack_dbn_from_file("GLBX-NQ-2025-01-01-2025-12-10/glbx-mdp3-20251201-20251209.ohlcv-1m.dbn.zst")
    print_candles(db_url)
    # request_data("NQ", "2025-11-01", "2025-12-01")
    # get_dbn_historical_batch("NQ", "2025-01-01", "2025-12-10")


if __name__ == "__main__":
    main()
