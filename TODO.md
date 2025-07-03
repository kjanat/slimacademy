# SlimAcademy Enhancement TODO

## Implementation Progress

### 🔒 Phase 1: Security & Sanitization Overhaul (CRITICAL)

#### 1.1 HTML Escaping Enhancement
- [x] **COMPLETED** - Fix XSS vulnerability in `internal/writers/html.go:369` using proper URL validation
- [x] **COMPLETED** - Replace basic HTML escaping (`html.go:410-417`) with `html/template.HTMLEscapeString`
- [x] **COMPLETED** - Add context-aware escaping for attributes vs content
- [x] **COMPLETED** - Add secure URL validation for href and img src attributes

#### 1.2 URL Security
- [x] **COMPLETED** - Add `net/url` validation with scheme allowlisting (`http`, `https`, `mailto`, `tel`)
- [x] **COMPLETED** - Replace dangerous protocols (`javascript:`, `data:`, `vbscript:`) with safe defaults
- [x] **COMPLETED** - Enhance `internal/sanitizer/sanitizer.go` URL cleaning with comprehensive security validation

#### 1.3 Input Validation
- [ ] Add JSON schema validation at parser entry point
- [ ] Implement comprehensive sanitization before processing
- [ ] Add CSP-compatible output (strip inline CSS/JS)

### 🏗️ Phase 2: Architecture Modernization

#### 2.1 HAST Integration
- [x] **COMPLETED** - Create `internal/hast/` package with enhanced HAST types and utilities
- [x] **COMPLETED** - Implement HAST→HTML pipeline with secure rendering
- [x] **COMPLETED** - Add EventToHASTConverter for streaming events to HAST transformation
- [x] **COMPLETED** - Include comprehensive test coverage and visitor pattern support

#### 2.2 Layered Architecture
- [ ] Implement enhanced JSON parsing with validation
- [ ] Create Google Docs → HAST conversion layer
- [ ] Add post-processing optimizations
- [ ] Upgrade output layer with template system

#### 2.3 Writer Registry Enhancement
- [ ] Upgrade `internal/writers/registry.go` with HAST support
- [ ] Add proper error propagation and context cancellation
- [ ] Implement concurrent processing with safety guarantees

### ⚡ Phase 3: Go 1.24 Feature Integration

#### 3.1 Tool Directive Migration
- [x] **COMPLETED** - Migrate to `go.mod` tool directive (no existing tools.go found)
- [x] **COMPLETED** - Add development tools: `golangci-lint`, `staticcheck`, `gosec`, `govulncheck`
- [x] **COMPLETED** - Create reproducible development environment with Makefile

#### 3.2 Generic Type Safety
- [ ] Replace `interface{}` usage with generic type aliases
- [ ] Create `type Stack[T any]` for formatting state management
- [ ] Add type-safe event handling with generics

#### 3.3 Advanced Iterators
- [ ] Implement `bytes.Lines()` iterators for memory-efficient processing
- [ ] Use `iter.Seq[Event]` patterns from Go 1.23+
- [ ] Add streaming with `io.Pipe` between parser and writer

#### 3.4 Memory Management
- [ ] Add `sync.Pool` for `bytes.Buffer` instances
- [ ] Implement `weak` package for caching with automatic cleanup
- [ ] Use Go 1.24's improved allocator and Swiss-table maps

### 🚀 Phase 4: Performance Optimization

#### 4.1 Streaming Architecture
- [ ] Replace full document copies with streaming pipeline
- [ ] Implement O(1) memory usage for large documents
- [ ] Add proper context cancellation and cleanup

#### 4.2 Resource Management
- [ ] Embed CSS with `//go:embed` directive
- [ ] Use `os.Root` for secure file operations
- [ ] Implement efficient slug generation with collision handling

#### 4.3 Concurrent Processing
- [ ] Enhance MultiWriter with context-aware processing
- [ ] Add proper error aggregation and cancellation
- [ ] Implement pipeline parallelism where safe

### 🧪 Phase 5: Testing & Quality Infrastructure

#### 5.1 Fuzzing Integration
- [x] **COMPLETED** - Add `go test -fuzz=FuzzSanitize` for security testing
- [x] **COMPLETED** - Implement comprehensive fuzzing tests for sanitizer
- [x] **COMPLETED** - Create comprehensive edge case coverage with property-based testing

#### 5.2 Advanced Testing
- [ ] Use `testing/synctest` for concurrency testing
- [ ] Implement golden file testing patterns from slim project
- [ ] Add round-trip validation tests

#### 5.3 Static Analysis Pipeline
- [ ] Set up `go vet ./...` with tests analyzer
- [ ] Configure `staticcheck ./...` for advanced static analysis
- [ ] Add `gosec ./...` for security vulnerability scanning
- [ ] Implement `govulncheck ./...` for known vulnerability detection

### 📄 Phase 6: Template System Enhancement

#### 6.1 Template Integration
- [ ] Port `pkg/html/template.go` responsive template system
- [ ] Add tabbed interface with search functionality
- [ ] Implement academic metadata display

#### 6.2 Enhanced Styling
- [ ] Embed CSS with modern responsive design
- [ ] Add accessibility improvements
- [ ] Implement print-friendly layouts

## Current Focus

**CRITICAL SECURITY FIXES** - Working on XSS vulnerability in HTML writer

## Notes

- Implementation started: 2025-01-03
- Current priority: Fix critical security vulnerabilities first
- All changes maintain backward compatibility
- Comprehensive testing required for each phase

---

*Last updated: 2025-01-03*
