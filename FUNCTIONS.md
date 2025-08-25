# PMFS Functions

This document lists all exported functions in the PMFS package and what they do.

## LoadSetup
Sets up the base directory, ensures required folders exist and loads `index.toml`.

## (*Database) AddProduct
Appends a product to the database and creates its directory skeleton.

## (*Database) Save
Writes the in-memory index to `index.toml`.

## (*ProductType) AddProject
Adds a new project to a product and persists the change to disk.

## (*ProjectType) SaveProject
Writes the project's data to its `project.toml` file.

## (*ProjectType) LoadProject
Loads a project's data from its `project.toml` file.

## (*ProductType) LoadProjects
Loads all projects for a given product.

## (*Index) LoadAllProjects
Loads all projects for all products in the index.

## (*ProjectType) IngestInputDir
Scans a directory and ingests each file as an attachment.

## (*ProjectType) AddAttachmentFromInput
Moves a single file into the project's attachments and records minimal metadata.

## FromGemini
Converts a Gemini requirement into a PMFS requirement by copying its name and description.

