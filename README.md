# Wonk
Personal Htmx app

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
There are many ways to install sqlite3, feel free to research for your environment.

## How to run
To run the server run the following command:
```bash
go run cmd/main.go
```
## Makefile
When changes are made to the templ files we need to generate new _templ.go and tailwind files.
To help with this we run this command which uses makefile:
```bash
make runw
```
