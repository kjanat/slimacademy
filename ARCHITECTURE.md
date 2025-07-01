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

## Event-Driven Core

The core of the application is an event-driven system that processes the book's content as a stream of events. The `internal/events/event.go` file defines the structure of the events and the event stream.

### Events

The `Event` struct represents a single structural or formatting event in the document. It has the following fields:

*   **`Kind`**: The type of event, such as `StartDoc`, `EndDoc`, `StartParagraph`, `EndParagraph`, etc.
*   **`Arg`**: The data associated with the event, such as the text of a paragraph, the level of a heading, or the URL of an image.

The `Kind` of an event is represented by an integer constant, which allows for efficient processing of events.

### Event Stream

The event stream is a sequence of events that represents the entire content of the book. The application generates the event stream by traversing the `models.Book` struct and creating events for each element in the book.

## Writers

The `internal/writers/` directory contains a series of writers, each responsible for generating a specific output format. The following writers are currently implemented:

*   **`epub.go`**: Generates an EPUB file.
*   **`html.go`**: Generates an HTML file.
*   **`latex.go`**: Generates a LaTeX file.
*   **`markdown.go`**: Generates a Markdown file.
*   **`plaintext.go`**: Generates a plain text file.

Each writer consumes the event stream and generates the corresponding output format. The writers are designed to be independent of each other, which allows for concurrent processing of multiple output formats.

## Concurrency

The `internal/exporters/multi.go` file provides a mechanism for concurrently processing multiple output formats. The `MultiExporter` struct takes a list of writers and an event stream, and it concurrently feeds the events to each writer. This allows for efficient generation of multiple output formats from a single input source.

## Configuration

The `internal/config/` directory contains the configuration for the different output formats. The configuration files allow for customization of the output, such as the style of the HTML, the formatting of the LaTeX, and the structure of the EPUB.
