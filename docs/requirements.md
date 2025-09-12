# Requirements

This document outlines the key functional requirements for the PMFS library.

## Data Storage
- Maintain a file-based database rooted at a configurable base directory.
- Store products and projects using `index.toml` and per-project `project.toml` files.

## Product and Project Management
- Create, modify and persist products and projects.
- Load existing data from disk and expose in-memory models for manipulation.

## Attachment Ingestion
- Ingest files from an input folder, move them into project attachments and
  analyse them with a configured LLM client.
- Support text attachments created directly from strings.

## Requirement Handling
- Add, activate and deduplicate requirements within a project.
- Generate related requirements and design aspects using LLM assistance.

## Quality Control
- Evaluate requirements against configurable quality gates powered by the LLM.
- Scan entire projects and record gate results and follow‑up information.

## Excel Interoperability
- Export project data to Excel workbooks and import data back into the project,
  merging new and existing requirements.

## LLM Integration
- Interact with Gemini via a pluggable client with configurable model and
  request rate limits.
- Provide prompts for multiple roles and quality gates.

