# About budva43 Project

## Description

**budva43** is an intelligent automatic message forwarding system for Telegram, written in Go. The project represents an enterprise-level solution implementing UNIX-way principles and clean architecture to create thematic digests from messages from various channels and groups.

## Core Functionality

### Automatic Message Forwarding
- **Forward** — sends messages preserving original authorship
- **Send Copy** — creates message copies without indicating the original source
- **Media Album Support** — processing grouped images and files

### Filtering System
- **Exclusion Filters** — regular expressions to block unwanted content
- **Inclusion Filters** — rules to pass only relevant messages
- **Substring Filters** — precise filtering by substrings using regex groups
- **Auto Answers** — automatic responses to messages with keyboards
- **Special Chats** — automatic sending of filtered messages to check/other channels

### Content Transformation
- **Link Replacement** — automatic replacement of links to own messages in target chats
- **External Link Removal** — cleaning links to external sources
- **Text Fragment Replacement** — customizable text transformation for different recipients
- **Source Signatures** — indicating the original message source
- **Source Link Generation** — automatic addition of links to original messages

### Message Lifecycle Management
- **Copy Once** — single sending without synchronization on editing
- **Indelible** — protecting messages from deletion when original is deleted
- **Edit Synchronization** — automatic updating of copied messages when original changes
- **Delete Synchronization** — deleting copies when source message is deleted

### Additional Features
- **Rate Limiting** — sending speed control to prevent blocks
- **System Message Processing** — automatic deletion of service notifications

## Architecture

### Microservice Structure
The project is divided into two main services:

#### Engine (cmd/engine)
- **Purpose**: message forwarding execution
- **Restrictions**: sending new messages to outgoing chats is prohibited
- **Components**: Telegram update handlers, forwarding and filtering services

#### Facade (cmd/facade)  
- **Purpose**: providing APIs (GraphQL, gRPC, REST)
- **Capabilities**: full access to message sending functions
- **Interfaces**: web interface, gRPC API, terminal interface

### Layered Architecture

```
Transport Layer    → HTTP, gRPC, Terminal, Telegram Bot API
Service Layer      → Business logic, forwarding rules processing
Repository Layer   → TDLib, Storage, Queue
Domain Layer       → Data models, forwarding rules
```

### Design Patterns
- **Clean Architecture** — clear separation of responsibility layers
- **Dependency Injection** — custom dependency injection system
- **Repository Pattern** — data access abstraction
- **Observer Pattern** — Telegram update handling

## Technology Stack

### Core Technologies
- **Go 1.24** — main development language with generics support
- **TDLib** — official Telegram library for client applications
- **BadgerDB** — embedded NoSQL database

### API & Transport
- **gRPC** — high-performance API for integrations
- **GraphQL** — flexible API for web clients  
- **REST** — classic HTTP API
- **Telegram Client API** — direct interaction with Telegram
- **Terminal Interface** — interactive terminal interface

### Development & Testing
- **Docker & DevContainers** — development containerization
- **Testcontainers** — integration testing (including Redis for connection testing)
- **Mockery** — mock object generation
- **Godog (BDD)** — behavior-driven testing
- **GitHub Actions CI** — automated continuous integration

### Monitoring & Observability
- **Structured Logging** — structured logging with slog
- **Grafana + Loki** — centralized logs and monitoring
- **pplog** — human-readable JSON logging for development
- **spylog** — log interception in tests

### Build & Development Tools
- **Task** — Make alternative for task automation
- **golangci-lint** — comprehensive code quality checking
- **Custom Linters** — "error-log-or-return" and "unused-interface-methods"
- **protobuf** — gRPC interface generation
- **jq** — real-time log viewing with filtering
- **EditorConfig** — editor settings consistency

## Development Principles

### Architectural Principles
- **SOLID** — applying all five OOP principles
- **DRY** — avoiding code duplication without fanaticism
- **KISS** — preferring simple solutions over complex ones
- **YAGNI** — implementing only necessary functionality

### Go-Specific Approaches
- **CSP (Communicating Sequential Processes)** — using channels instead of mutexes
- **Interface Segregation** — local interfaces in consuming modules
- **Accept interfaces, return structs** — idiomatic interface work
- **Early Return** — reducing code nesting
- **Table-driven tests** — structured testing

### Error Handling Conventions
- **Structured Errors** — structured errors with automatic call stack
- **Log or Return** — either log error or return it up
- **Minimal Wrapping** — wrapping errors only when adding context

## Configuration

### Settings Hierarchy
```
defaultConfig() → config.yml → .env
```

### Configuration Types
- **Static Configuration** — basic application settings
- **Dynamic Configuration** — forwarding rules with hot-reload
- **Secret Data** — API keys and tokens via environment variables

### Forwarding Configuration Examples
```yaml
forward_rules:
  rule1:
    from: 1001234567890
    to: [1009876543210, 1001111111111]
    send_copy: true
    exclude: "EXCLUDE|spam"
    include: "IMPORTANT|urgent"
    copy_once: false
    indelible: true
```

## Testing

### Multi-Level Testing
- **Unit Tests** — isolated component testing with mocks
- **Integration Tests** — component interaction testing
- **E2E Tests** — complete user scenarios via gRPC API
- **BDD Tests** — behavior description in natural language
- **Snapshot Testing** — testing with reference snapshots

### Special Techniques
- **Synctest** — time and concurrency testing
- **Call-driven Tests** — table tests with preparation functions
- **Spy Logging** — log interception and verification in tests

### Test Coverage
- **Codecov.io Integration** — automatic code coverage tracking
- **Coverage for Integration Tests** — special tools
- **Functional Coverage** — all key usage scenarios
- **Technical Coverage** — internal functions and edge cases
- **BDD Scenarios** — user stories and business rules

## Deployment and Operations

### Launch Options
- **Local Development** — direct TDLib installation on host machine
- **DevContainer** — fully isolated development environment
- **Production** — containerized deployment

### Monitoring and Debugging
- **Structured Logs** — JSON logs for machine processing
- **Human-readable Logs** — pplog for development
- **Health Checks** — service status checks
- **Graceful Shutdown** — correct work termination

### Integrations
- **Telegram Client** — full-featured client with authorization
- **External APIs** — integration via GraphQL, gRPC, REST
- **Message Queues** — asynchronous message processing

## Contributing to Development

### Project Philosophy
Budva43 is positioned as "my best learning project for applying technologies — from MVP to Enterprise level". The project demonstrates modern Go development approaches, including the latest language features and industry best practices.

### Unique Features
- **Experimental Approaches** — using cutting-edge Go capabilities
- **Comprehensive Testing** — full spectrum of testing techniques
- **Production-ready Quality** — readiness for industrial use
- **Educational Character** — rich documentation and examples

The project is actively developed and serves as a demonstration of modern Go capabilities in enterprise development. 