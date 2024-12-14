# Wonk
Personal Htmx app

# Table Of Contents
- [Development](#development)

# Development
This app uses htmx, templ and tailwind

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
