# Change these variables as necessary.
main_package_path = ./cmd/main.go
binary_name = wonk
input_css_file = static/css/input.css
output_css_file = static/css/output.css

# --- Tailwind ---
## Tailwind watchmode
.PHONY: tailw
tailw:
	# WatchMode
	./tailwindcss -i ${input_css_file} -o ${output_css_file} --watch
.PHONY: tailo
tailo:
	# Production output
	./tailwindcss -i ${input_css_file} -o ${output_css_file} --minify

# --- Templ ---
## Templ generation
.PHONY: tgen
tgen:
	templ generate

# --- Development
.PHONY: runc
runc:
	go run ${main_package_path} -logfmt=color

## Gen Templ, Gen Tailwind, run api with color logs
.PHONY: runw
runw:
	templ generate
	./tailwindcss -i ${input_css_file} -o ${output_css_file}
	go run ${main_package_path} -logfmt=color

