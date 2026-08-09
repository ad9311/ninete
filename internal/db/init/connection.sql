-- Connection-level PRAGMAs. SQLite scopes every setting below to a single
-- connection and resets it to the built-in default on the next one, so these
-- run through the driver's connect hook for every connection the pool opens.
--
-- "foreign_keys" is the critical one: SQLite defaults it to OFF, and a
-- connection without it silently skips ON DELETE CASCADE, leaving orphaned
-- rows behind with no error.
PRAGMA foreign_keys = ON;
PRAGMA ignore_check_constraints = OFF;
PRAGMA recursive_triggers = ON;
PRAGMA trusted_schema = OFF;
PRAGMA synchronous = NORMAL;
PRAGMA temp_store = MEMORY;
PRAGMA cache_size = -5120;
PRAGMA temp.cache_size = -5120;
PRAGMA wal_autocheckpoint = 1000;
PRAGMA busy_timeout = 5000;
PRAGMA mmap_size = 67108864;
