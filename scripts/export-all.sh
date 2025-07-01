#!/bin/bash

# Find all directories in source/ (assuming each directory is a book)
# Use proper array handling to preserve directory names with spaces
mapfile -t books < <(find source/ -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)
formats=("html" "markdown" "epub")

# export_book exports a single book in the specified format by running the Go transformer program.
export_book() {
    local book="$1"
    local format="$2"
    echo "Exporting '$book' as $format..."
    go run ./cmd/transformer/main.go -book "$book" -format "$format"
    echo "Completed '$book' as $format"
}

# export_job prepares the export_book function for use in a subshell and invokes it with the specified book and format.
export_job() {
    export -f export_book
    export_book "$1" "$2"
}

# Start all export jobs in parallel
pids=()
for book in "${books[@]}"; do
    for format in "${formats[@]}"; do
        export_job "$book" "$format" &
        pids+=($!)
    done
done

echo "Started ${#pids[@]} export jobs in parallel..."

# Wait for all background jobs to complete
for pid in "${pids[@]}"; do
    wait $pid
done

echo "All exports completed!"
