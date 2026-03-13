import sqlite3
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
OUTPUT_DIR = ROOT / "testdata"
DB_PATH = OUTPUT_DIR / "mock-sqlite.db"


def create_tables(cursor: sqlite3.Cursor) -> None:
    cursor.executescript(
        """
        DROP TABLE IF EXISTS order_items;
        DROP TABLE IF EXISTS orders;
        DROP TABLE IF EXISTS customers;

        CREATE TABLE customers (
            id INTEGER PRIMARY KEY,
            full_name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            city TEXT NOT NULL,
            created_at TEXT NOT NULL
        );

        CREATE TABLE orders (
            id INTEGER PRIMARY KEY,
            customer_id INTEGER NOT NULL,
            order_number TEXT NOT NULL UNIQUE,
            status TEXT NOT NULL,
            total_amount REAL NOT NULL,
            ordered_at TEXT NOT NULL,
            FOREIGN KEY (customer_id) REFERENCES customers(id)
        );

        CREATE TABLE order_items (
            id INTEGER PRIMARY KEY,
            order_id INTEGER NOT NULL,
            product_name TEXT NOT NULL,
            quantity INTEGER NOT NULL,
            unit_price REAL NOT NULL,
            FOREIGN KEY (order_id) REFERENCES orders(id)
        );
        """
    )


def seed_data(cursor: sqlite3.Cursor) -> None:
    customers = [
        (1, "Alice Johnson", "alice@example.com", "Bangkok", "2026-01-02 09:00:00"),
        (2, "Ben Carter", "ben@example.com", "Chiang Mai", "2026-01-03 10:15:00"),
        (3, "Chloe Martin", "chloe@example.com", "Phuket", "2026-01-04 11:30:00"),
        (4, "Daniel Lee", "daniel@example.com", "Khon Kaen", "2026-01-05 12:45:00"),
        (5, "Emma Brown", "emma@example.com", "Pattaya", "2026-01-06 14:00:00"),
        (6, "Finn Walker", "finn@example.com", "Hat Yai", "2026-01-07 15:10:00"),
        (7, "Grace Hall", "grace@example.com", "Udon Thani", "2026-01-08 16:20:00"),
        (8, "Henry Young", "henry@example.com", "Nakhon Ratchasima", "2026-01-09 17:30:00"),
        (9, "Ivy King", "ivy@example.com", "Surat Thani", "2026-01-10 18:40:00"),
        (10, "Jack Scott", "jack@example.com", "Ayutthaya", "2026-01-11 19:50:00"),
    ]

    orders = [
        (1, 1, "ORD-1001", "paid", 120.50, "2026-02-01 08:30:00"),
        (2, 2, "ORD-1002", "pending", 89.99, "2026-02-02 09:00:00"),
        (3, 3, "ORD-1003", "shipped", 45.00, "2026-02-03 09:30:00"),
        (4, 4, "ORD-1004", "paid", 210.75, "2026-02-04 10:00:00"),
        (5, 5, "ORD-1005", "cancelled", 15.99, "2026-02-05 10:30:00"),
        (6, 6, "ORD-1006", "paid", 67.49, "2026-02-06 11:00:00"),
        (7, 7, "ORD-1007", "processing", 150.00, "2026-02-07 11:30:00"),
        (8, 8, "ORD-1008", "shipped", 99.95, "2026-02-08 12:00:00"),
        (9, 9, "ORD-1009", "paid", 34.25, "2026-02-09 12:30:00"),
        (10, 10, "ORD-1010", "pending", 500.00, "2026-02-10 13:00:00"),
    ]

    order_items = [
        (1, 1, "Wireless Mouse", 2, 25.00),
        (2, 2, "USB-C Cable", 3, 9.99),
        (3, 3, "Notebook", 5, 9.00),
        (4, 4, "Mechanical Keyboard", 1, 210.75),
        (5, 5, "Pen Set", 2, 8.00),
        (6, 6, "Laptop Stand", 1, 67.49),
        (7, 7, "Webcam", 2, 75.00),
        (8, 8, "Desk Lamp", 1, 99.95),
        (9, 9, "Water Bottle", 1, 34.25),
        (10, 10, "Monitor 27 inch", 2, 250.00),
    ]

    cursor.executemany(
        "INSERT INTO customers (id, full_name, email, city, created_at) VALUES (?, ?, ?, ?, ?)",
        customers,
    )
    cursor.executemany(
        "INSERT INTO orders (id, customer_id, order_number, status, total_amount, ordered_at) VALUES (?, ?, ?, ?, ?, ?)",
        orders,
    )
    cursor.executemany(
        "INSERT INTO order_items (id, order_id, product_name, quantity, unit_price) VALUES (?, ?, ?, ?, ?)",
        order_items,
    )


def main() -> None:
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    connection = sqlite3.connect(DB_PATH)
    try:
        cursor = connection.cursor()
        create_tables(cursor)
        seed_data(cursor)
        connection.commit()
    finally:
        connection.close()

    print(DB_PATH)


if __name__ == "__main__":
    main()
