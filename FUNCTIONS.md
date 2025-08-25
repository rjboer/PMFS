# PMFS Functions

This document lists all exported functions in the PMFS package and what they do.

## EnsureLayout
Creates base folder structure and ensures `index.toml` exists.

## Load
Reads `index.toml` into the in-memory database model.

## (*Database) NewProduct
Creates a product, writes the index, and returns its ID.

## (*Database) Save
Writes the in-memory database to `index.toml`.

## (*Product) CreateProject
Adds a new project to a product and persists the change to disk.

## (*Project) Save
Writes the project's data to its `project.toml` file.

## (*Project) Load
Loads a project's data from its `project.toml` file.

## (*Product) LoadProjects
Loads all projects for a given product.

## (*Database) LoadAllProjects
Loads all projects for all products in the database.

## (*Project) IngestInputDir
Scans a directory and ingests each file as an attachment.

## (*Project) AddAttachmentFromInput
Moves a single file into the project's attachments and records minimal metadata.

## FromGemini
Converts a Gemini requirement into a PMFS requirement by copying its name and description.

