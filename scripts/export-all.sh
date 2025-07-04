#!/bin/bash

# Find all directories in source/ (assuming each directory is a book)
# Use portable approach for directory listing
books=()
while IFS= read -r -d '' dir; do
    books+=("$(basename "$dir")")
done < <(find source/ -mindepth 1 -maxdepth 1 -type d -print0)

formats=("html")
# formats=("html" "markdown" "epub")

# Build the slim binary once for efficiency
echo "Building slim binary..."
go build -o ./bin/slim ./cmd/slim
if [[ $? -ne 0 ]]; then
    echo "Failed to build slim binary"
    exit 1
fi

# export_book exports a single book in the specified format using the pre-built slim binary.
export_book() {
    local book="$1"
    local format="$2"
    local output_file="outputs/${book}.${format}"
    echo "Exporting '${book}' as ${format}..."

    # Create outputs directory if it doesn't exist
    mkdir -p outputs

    # Use correct CLI syntax: input as positional argument, output as flag
    ./bin/slim convert --format "${format}" "source/${book}" --output "${output_file}"

    if [[ $? -eq 0 ]]; then
        echo "Completed '${book}' as ${format} -> ${output_file}"
    else
        echo "Failed '${book}' as ${format}"
    fi
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
        export_job "${book}" "${format}" &
        pids+=($!)
    done
done

echo "Started ${#pids[@]} export jobs in parallel..."

# Wait for all background jobs to complete
for pid in "${pids[@]}"; do
    wait "${pid}"
done

echo "All exports completed!"
