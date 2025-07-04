# Architecture

This document provides a comprehensive overview of the architecture of the Slim Academy project.

## High-Level Overview

The Slim Academy project is a Go-based application that converts book content from a structured JSON format into various output formats, including Markdown, HTML, LaTeX, and EPUB. The application is designed to be extensible, allowing for the addition of new output formats with minimal effort.

The architecture is based on an event-driven model, where the input JSON is parsed into a stream of events that represent the structure and content of the book. These events are then consumed by a series of writers, each responsible for generating a specific output format. This decoupled design allows for concurrent processing of multiple output formats and ensures that the core logic of the application remains independent of the output format.

## Input Format

The application uses a structured JSON format as its input. The JSON data is organized into a directory containing several files, including:

*   **`chapters.json`**: Defines the structure of the book's chapters and subchapters.
*   **`content.json`**: Contains the main content of the book, including paragraphs, headings, lists, tables, and images.
*   **`*.json`**: A metadata file that contains information about the book, such as its title, description, and other relevant details.

The JSON data is parsed into a `models.Book` struct, which serves as the in-memory representation of the book's content.

## Parsing

The `internal/parser/json_parser.go` file is responsible for parsing the input JSON files into a `models.Book` struct. The `BookParser` struct provides the following methods:

*   **`ParseBook(bookDirPath string)`**: Parses the JSON files in the specified directory and returns a `models.Book` struct.
*   **`FindAllBooks(rootDir string)`**: Traverses a directory and finds all the book directories.

The parser first reads the metadata file, then the `chapters.json` file, and finally the `content.json` file. The parsed data is then used to populate the fields of the `models.Book` struct.

## Event-Driven Streaming Core

The core of the application is a modernized event-driven streaming system that processes the book's content as a memory-efficient stream of events. The `internal/streaming/stream.go` file defines the structure of the events and implements Go 1.23+ iterators for O(1) memory usage.

### Events

The `Event` struct represents a single structural or formatting event in the document with concrete typed fields:

*   **`Kind`**: The type of event, such as `StartDoc`, `EndDoc`, `StartParagraph`, `EndParagraph`, etc.
*   **Concrete Fields**: Direct access to typed data like `Title`, `HeadingText`, `TextContent`, `ImageURL`, `Style`, etc.
*   **Memory Optimization**: Uses `unique.Handle[string]` for O(1) duplicate detection and interning.

The events use concrete structs instead of `any` interface for better type safety and performance.

### Event Stream

The event stream is a sequence of events generated using Go 1.23+ `iter.Seq[Event]` iterators. The streaming system includes:
*   **Token Sanitization**: Content is sanitized before streaming with diagnostic reporting
*   **Memory Efficiency**: O(1) RAM usage even on large documents using `bytes.Lines()` for chunking
*   **Context Cancellation**: Proper cancellation and error propagation support

## Writers

The `internal/writers/` directory contains a series of writers, each responsible for generating a specific output format. The following writers are currently implemented:

*   **`epub.go`**: Generates an EPUB file.
*   **`html.go`**: Generates an HTML file.
*   **`latex.go`**: Generates a LaTeX file.
*   **`markdown.go`**: Generates a Markdown file.
*   **`plaintext.go`**: Generates a plain text file.

Each writer consumes the event stream and generates the corresponding output format. The writers are designed to be independent of each other, which allows for concurrent processing of multiple output formats.

## Writer Registry and Concurrency

The `internal/writers/registry.go` file provides a modernized mechanism for concurrently processing multiple output formats. The new system includes:

*   **Auto-Registration**: Writers automatically register themselves via `init()` functions
*   **WriterV2 Interface**: Enhanced interface with error handling, statistics, and proper resource management
*   **MultiWriter**: Context-aware concurrent processing with error propagation
*   **Shared Stack Logic**: Common structural stack management for tables, lists, and formatting in `stack.go`

The registry system takes registered writers and an event stream, concurrently feeding events to each writer through the MultiWriter implementation. This allows for efficient generation of multiple output formats from a single input source with proper error handling and observability.

## Configuration

The `internal/config/` directory contains the configuration for the different output formats. The configuration files allow for customization of the output, such as the style of the HTML, the formatting of the LaTeX, and the structure of the EPUB.
