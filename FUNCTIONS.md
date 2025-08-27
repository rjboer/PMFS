# PMFS Functions

This document lists all exported functions in the PMFS package and what they do.

## LoadSetup
Sets up the base directory, ensures required folders exist and loads `index.toml`.

## (*Database) NewProduct
Creates a product, writes it to `index.toml` and returns its ID.

## (*Database) ModifyProduct
Updates an existing product and persists the change to `index.toml`.

## (*Database) Save
Writes the in-memory database to `index.toml`.

## (*ProductType) NewProject
Adds a new project to a product, writes its `project.toml` and updates the index.

## (*ProductType) Project
Loads and returns a specific project by ID.

## (*ProjectType) Save
Writes the project's data to its `project.toml` file.

## (*ProjectType) Load
Loads a project's data from its `project.toml` file.

## (*ProductType) LoadProjects
Loads all projects for a given product.

## (*Database) LoadAllProjects
Loads all projects for all products in the database.

## (*ProjectType) IngestInputDir
Scans a directory and ingests each file as an attachment.

## (*ProjectType) AddAttachmentFromInput
Moves a single file into the project's attachments and records minimal metadata.

## FromGemini
Converts a Gemini requirement into a PMFS requirement by copying its name and description.

