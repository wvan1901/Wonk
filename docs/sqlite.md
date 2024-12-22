# Sqlite
This is a doc to help remember commands

# How to install cli
How to install sqlite3 cli tool
## Mac
```bash
brew install sqlite3
brew install sqlite-utils
```

# Useful commands
`NOTE:` replace `{...}` with the desired name.
## Create or Access Database
```bash
sqlite3 {SOME_DB}.db
```

## Run sql file
```bash
sqlite3 {SOME_DB}.db < {PATH_TO_SQL_FILE}.sql
```

## Run sql query
```bash
sqlite-utils {SOME_DB}.db {SOME_SQL_QUERY} --table
```
