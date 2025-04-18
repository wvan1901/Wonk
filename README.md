# Wonk
This is my personal Htmx app.

# Table Of Contents
- [Development](#development)

# Development
This app uses htmx, templ and tailwind

## Set up
### Env file
A `.env` file is required with the following:
```bash
# Must be a vaild hex string
COOKIE_SECRET_KEY=""
JWT_SECRET_KEY=""
```

### Templ
Follow thier docs for installation steps: [Docs](https://templ.guide/quick-start/installation)

### Tailwind
I perfer to use tailwinds standalone cli to run tailwind but other options are available.
Docs for stand alone cli: [Docs](https://tailwindcss.com/blog/standalone-cli)

### Sqlite3
There are many ways to install sqlite3, feel free to research for your environment.\
Once you have sqlite installed create a db file called `wonk.db`. This file serves as the database for the application.
Once the `wonk.db` file is created run the scripts in the folder `sqlite/scripts/`, once completed the db is ready to be used.\
Below are commands to help:
```bash
# Create db file
sqlite3 wonk.db
# Run sql file in db
sqlite3 wonk.db < sqlite/scripts/createTables.sql
```

## How to run
### Generate Templ Files
When modifing templ files we need to generate their output go files, we can do that with the following command:
```bash
templ generate
```

### Generate Tailwind file
When we add a new tailwind class we need the css to be updated, we can do that with the following command:
```bash
./tailwindcss -i static/css/input.css -o static/css/output.css
```

### Run App
To run the server run the following command:
```bash
go run cmd/main.go
# NOTE: I use this command to run my own logger (There is no need, but I like colored logs)
go run ./cmd/main.go -logfmt=devlog
```

### Makefile
Remembering and running all the commands above can be cumbersome, to fix this I use a makefile.\
This command will Generate the templ files, generate the tailwind file, and run the app.
```bash
make runw
```
