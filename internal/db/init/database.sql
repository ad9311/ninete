-- Database-level PRAGMAs. SQLite persists these in the database file itself, so
-- they only need to run once when the pool is opened. Per-connection settings
-- belong in "connection.sql" instead.
PRAGMA encoding = "UTF-8";
PRAGMA page_size = 4096;
PRAGMA auto_vacuum = INCREMENTAL;
PRAGMA application_id = 0x6E696E657465;
PRAGMA journal_mode = WAL;
PRAGMA optimize;
