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
  analyze them with a configured LLM client.
- Support text attachments created directly from strings.
- Store extracted "intelligence" from each attachment and use it to
  optionally generate design aspects and candidate requirements.

## Requirement Handling
- Add, activate, delete, restore and deduplicate requirements within a project.
- Generate related requirements and design aspects using LLM assistance.

## Quality Control
- Evaluate requirements against configurable quality gates powered by the LLM.
- Define processing scope rules:
  - **QualityControlPending** processes active requirements that have not yet
    been analyzed.
  - **AnalyzeAll** processes all non-proposed, non-deleted requirements,
    regardless of prior analysis.
- Record gate results and follow-up information.

## Excel Interoperability
- Export project data to Excel workbooks and import data back into the project.
- Importing must support merge and replace modes.
  - **Merge** updates requirements by ID and appends new ones.
  - **Replace** discards all current requirements before importing.
  - IDs from the spreadsheet are preserved; missing IDs are assigned during import.

## LLM Integration
- Interact with Gemini via a pluggable client with configurable model and
  request rate limits.
- Provide prompts for multiple roles and quality gates.

## Requirement Lifecycle
- Track requirement state with flags: `Proposed`, `AIgenerated`, `AIanalyzed`,
  `Active`, and `Deleted`.
- Allowed transitions include:
  - `Proposed` → `Active` via activation.
  - `Proposed` or `Active` → `Deleted` via deletion.
  - `Deleted` → `Active` via restoration.
- Batch operations skip `Proposed` and `Deleted` requirements unless specified.
- Deleted requirements remain in exports and imports with their `Deleted` flag preserved.

## Naming Conventions
- Public APIs use American English spelling (e.g., `Analyze`).

