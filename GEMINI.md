# Gemini Code Assistant: Project Overview - Dirty Congress

This document provides a high-level overview of the "Dirty Congress" application architecture, as analyzed by the Gemini Code Assistant.

## 1. Project Summary

The "Dirty Congress" project is a Go-based monolithic web application designed to track, analyze, and display data related to the U.S. Congress. It aggregates information about congress members, bills, voting records, lobbying, and officially sponsored travel, presenting it through a web interface.

The application has two primary modes of operation:
1.  **Web Server:** A public-facing website for users to explore the collected data.
2.  **CLI Tools:** A suite of command-line utilities for data ingestion, processing, and database management.

## 2. Architecture: Modular Monolith

The application is structured as a **modular monolith**. While it is deployed as a single application, the codebase is organized into distinct, feature-oriented packages. This approach combines the simplicity of a monolithic deployment with the organizational benefits of a microservices architecture.

- **Core Logic:** The main application logic resides in the `internal/` directory.
- **Feature Packages:** Each major feature (e.g., `congress`, `travel`, `stocks`) is encapsulated in its own package within `internal/`. These packages typically contain their own HTTP route handlers, database models, and UI templates.
- **Shared Kernel:** A central package, `internal/m/`, provides shared code for data models (GORM), the templating engine, and other utilities, serving as a "kernel" for the different feature modules.

## 3. Technology Stack

- **Backend:**
    - **Language:** Go
    - **Web Framework:** [Fiber](https://gofiber.io/) (An Express.js-inspired web framework for Go)
    - **ORM:** [GORM](https://gorm.io/) for database interaction.
- **Frontend:**
    - **Templating:** [Handlebars](https://handlebarsjs.com/) is used for server-side rendering of HTML.
    - **Dynamic UI:** [HTMX](https://htmx.org/) is used to provide dynamic, client-side updates without writing complex JavaScript.
- **Database:**
    - The specific database is not explicitly defined in the core configuration, but the use of GORM allows for flexibility. The presence of Ansible scripts suggests a production deployment likely uses a robust database like PostgreSQL, while local development might use SQLite.
- **AI & Data Processing:**
    - **PDF Analysis:** The application uses Google's Vertex AI (Gemini models) to analyze and extract structured data from PDF documents, particularly for travel records.
- **Deployment & Automation:**
    - **Orchestration:** [Ansible](https://www.ansible.com/) is used for automating the build, deployment, and server setup processes.

## 4. Key Directory & File Structure

- `main.go`: The primary entry point for the entire application. It uses command-line flags to determine whether to run the web server or execute a specific data processing script.
- `internal/app/controllers.go`: This is the heart of the web server. It initializes the Fiber app, sets up middleware (like logging and recovery), and mounts the routes defined in the various feature packages.
- `internal/m/template_engine.go`: A crucial part of the frontend architecture. It configures the Handlebars templating engine and uses Go's `embed` feature to package all `.hbs` template files directly into the compiled application binary. This simplifies deployment significantly.
- `internal/m/models.go`: Defines the core database schema using GORM struct tags. It establishes the `SetupDB` function to configure the database connection.
- `cmd/`: Contains the entry points for the various command-line tools used for data ingestion and processing (e.g., `scrape-official_travel`).
- `*.ansible.yml` / `ansible.cfg`: Ansible playbooks and configuration files that define the infrastructure and deployment process, including server setup, application build, and service management.

## 5. Data Flow

1.  **Ingestion:** Data is collected from various sources using custom scripts located in the `cmd/` directory. This includes scraping websites, processing XML feeds (e.g., from the Federal Register), and analyzing PDF documents with AI.
2.  **Storage:** The processed data is stored in a relational database (likely PostgreSQL or SQLite) according to the GORM models defined in `internal/m/models.go` and other feature-specific model files.
3.  **Presentation:** The Fiber web server queries the database and renders the data into HTML using the Handlebars templates. HTMX is used to enhance the user experience with dynamic sorting, filtering, and loading of data without full page reloads.
