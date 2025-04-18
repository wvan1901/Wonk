# Wonk
This is my personal HTMX app.

# Table Of Contents
- [Development](#development)

# Development
This app uses HTMX, Templ, and Tailwind.
## Set up
### Env file
A `.env` file is required with the following:
```bash
# Must be a vaild hex string
COOKIE_SECRET_KEY=""
JWT_SECRET_KEY=""
```

### Templ
Follow their docs for installation steps: [Docs](https://templ.guide/quick-start/installation)

### Tailwind
I prefer to use Tailwind's standalone CLI to run Tailwind, but other options are available.
Docs for standalone CLI: [Docs](https://tailwindcss.com/blog/standalone-cli)

### SQLite3
There are many ways to install sqlite3; feel free to research for your environment.\
Once you have SQLite installed, create a DB file called `wonk.db`. This file serves as the database for the application.\
Once the `wonk.db` file is created, run the scripts in the folder `sqlite/scripts/`, once completed, the DB is ready to be used.\
Below are commands to help:
```bash
# Create db file
sqlite3 wonk.db
# Run sql file in db
sqlite3 wonk.db < sqlite/scripts/createTables.sql
```

## How To Run
### Generate Templ Files
When modifying templ files, we need to generate their output go files. It can be done with the following command:
```bash
templ generate
```

### Generate Tailwind file
When we add a new Tailwind class we need the CSS to be updated. It can be done with the following command:
```bash
./tailwindcss -i static/css/input.css -o static/css/output.css
```

### Run App
The following command runs the server:
```bash
go run cmd/main.go
# NOTE: I use this command to run my own logger (There is no need, but I like colored logs)
go run ./cmd/main.go -logfmt=devlog
```

### Makefile
Remembering and running all the commands above can be cumbersome, to fix this I use a Makefile.\
This command will Generate the Templ files, generate the Tailwind file, and run the server.
```bash
make runw
```
