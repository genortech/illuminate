# Illuminate Project Analysis

This document provides a summary of the `illuminate` project based on an analysis of its source code and configuration files.

## 1. Project Overview

`illuminate` is a web application written in Go. It uses a modern stack that combines a Go backend with a dynamic frontend powered by HTMX and styled with Tailwind CSS. The project is containerized using Docker and managed with Make.

Based on the project plan, the application is intended to be a **Photometric Lighting File Converter** that can process and convert between IES, LDT, and CIE formats.

## 2. Technology Stack

- **Backend:**
  - **Language:** Go (Version 1.25.1 or higher)
  - **Web Framework:** [Echo](https://echo.labstack.com/) (`github.com/labstack/echo/v4`)
  - **Templating Engine:** [Templ](https://templ.guide/) (`github.com/a-h/templ`) for Go components.
  - **Database:** [SQLite](https://www.sqlite.org/index.html) (`github.com/mattn/go-sqlite3`)

- **Frontend:**
  - **Styling:** [Tailwind CSS](https://tailwindcss.com/)
  - **Interactivity:** [HTMX](https://htmx.org/)

- **Tooling & DevOps:**
  - **Build/Task Runner:** `make`
  - **Containerization:** Docker (`docker-compose.yml`, `Dockerfile`)
  - **Live Reload:** [Air](https://github.com/air-verse/air) (`.air.toml`) for real-time development.
  - **Release Management:** [GoReleaser](https://goreleaser.com/) (`.goreleaser.yml`) for creating releases.
  - **CI/CD:** GitHub Actions (`.github/workflows/`) for testing and releases.

## 3. Project Structure

The project follows a standard Go application layout:

- `cmd/`: Contains the main entry points for the application.
  - `api/`: The main backend server application.
  - `web/`: Holds frontend assets like CSS, JavaScript (`htmx.min.js`), and Tailwind CSS configuration.
- `internal/`: Contains the core application logic, which is not meant to be imported by other projects.
  - `database/`: Manages the database connection (SQLite).
  - `server/`: Sets up the Echo web server and its routes.
- `Makefile`: Defines common development tasks like building, running, testing, and cleaning.
- `go.mod`: Manages the project's Go dependencies.
- `docker-compose.yml`: Defines the services for running the application with Docker, likely including the database.
- `.air.toml`: Configures the live-reloading environment for development.

## 4. How to Run the Project

The `Makefile` provides several commands to manage the application lifecycle:

- **Build the application:**

  ```bash
  make build
  ```

  This command generates Templ components, builds the Tailwind CSS, and compiles the Go binary.

- **Run the application:**

  ```bash
  make run
  ```

- **Run with Live Reload (Recommended for Development):**

  ```bash
  make watch
  ```

- **Run with Docker:**

  ```bash
  make docker-run
  ```

- **Run tests:**

  ```bash
  make test
  ```

---

## 5. Project Plan

This section is derived from the files found in the `project_plan` directory.

### 5.1. Requirements

**Functional:**

- **File Processing:** Convert between IES, LDT, and CIE lighting file formats.
- **Validation:** Validate files against standards (IES LM-63, EULUMDAT 1.0, etc.).
- **API:** Provide both a CLI for batch processing and a REST API for web integration.

**Non-Functional:**

- **Performance:** Process files up to 10MB in under 2 seconds.
- **Reliability:** Ensure data integrity and provide comprehensive error handling.
- **Usability:** Offer clear documentation for both CLI and API.

### 5.2. System Design

- **Architecture:** The system is designed using a modular, hexagonal architecture.
- **Data Flow:** `Input File` → `Format Detection` → `Parser` → `Validation` → `Common Model` → `Transformation` → `Target Writer` → `Output File`.
- **Interfaces:** Core logic is abstracted via `Parser` and `Writer` interfaces.
- **API Endpoints:**
  - `POST /api/v1/convert`: Single file conversion.
  - `POST /api/v1/batch`: Batch conversion.
- **CLI:** A `lighting-converter` CLI tool is planned for local conversions and validation.

### 5.3. Task Breakdown

The project is broken down into 10 phases:

1. **Project Setup & Foundation:** Environment and structure initialization.
2. **Core Data Models & Interfaces:** Define shared data structures and core interfaces.
3. **File Format Parsers:** Implement parsers for IES, LDT, and CIE.
4. **File Format Writers:** Implement writers for IES, LDT, and CIE.
5. **Conversion Logic & Orchestration:** Build the conversion manager and validation logic.
6. **CLI Application:** Develop the command-line interface.
7. **REST API Implementation:** Build the HTTP server and API endpoints.
8. **Testing & Quality Assurance:** Write unit and integration tests to achieve >90% coverage.
9. **Documentation & Deployment:** Finalize documentation and set up CI/CD pipelines.
10. **Release & Maintenance:** Versioning, release, and post-release monitoring.

The estimated total duration for the project is **25-35 days** for a single developer.
