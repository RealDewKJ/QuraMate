import duckdb

con = duckdb.connect('test_duckdb.db')
con.execute("CREATE TABLE test_table (id INTEGER, name VARCHAR)")
con.execute("INSERT INTO test_table VALUES (1, 'Hello DuckDB')")
con.execute("CREATE VIEW test_view AS SELECT * FROM test_table")
con.close()
