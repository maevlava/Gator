# Gator 🐊

Gator is a command-line tool for aggregating RSS and Atom feeds directly in your terminal. Follow your favorite blogs and news sources without needing a separate app. It fetches feeds in the background and stores posts locally in a PostgreSQL database.

## Prerequisites

Before you begin, ensure you have the following installed:

* **Go:** Version 1.18 or later is recommended. ([Installation Guide](https://go.dev/doc/install))
* **PostgreSQL:** A running instance that Gator can connect to. ([Download Page](https://www.postgresql.org/download/))
* **Git:** Required to fetch the code for installation.

## Installation

You can install the `gator` CLI directly using `go install`:

```bash
go install [github.com/](https://github.com/maevlava/Gator@latest