# Test Data

`mock-sqlite.db` is a sample SQLite database for connection testing in QuraMate.

Schema:

- `customers` with 10 rows
- `orders` with 10 rows
- `order_items` with 10 rows

Regenerate it with:

```powershell
python scripts/create_mock_sqlite.py
```
