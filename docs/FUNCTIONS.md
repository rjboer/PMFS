# Function Reference

This document lists exported functions across the PMFS repository and summarises their behaviour.

## Package `PMFS`

### SetBaseDir
Overrides the base data directory and refreshes internal paths.

### LoadSetup
Initialises the on-disk layout at the given path, loads the database and sets the default LLM client.

### (*Database) Save
Persists the in-memory database back to `index.toml`.

### (*Requirement) Analyse
Asks the configured LLM a role/question pair about the requirement's description.

### (*Requirement) EvaluateGates
Runs quality gates against the requirement using the configured LLM and stores the results.

### (*Requirement) QualityControlAI
Combines Analyse and EvaluateGates, marking the requirement as analysed.

### (*DesignAspect) EvaluateDesignGates
Evaluates gate checks for each template requirement in the design aspect.

### FromGemini
Converts a Gemini requirement value into a PMFS requirement structure.

### Deduplicate
Merges near-identical requirements, optionally ignoring proposed ones.

### (*Attachment) Analyze
Invokes the LLM to analyse the attachment and extract intelligence.

### (*Attachment) GenerateRequirements
Generates requirements from the attachment and appends them to the project.

### (*Attachment) Analyse
Runs a role/question pair against the attachment using the project's LLM client.

### (*Database) NewProduct
Creates a new product, writes it to `index.toml` and returns its ID.

### (*Database) ModifyProduct
Updates an existing product and persists the change to `index.toml`.

### (*ProductType) NewProject
Adds a project to the product, writes its `project.toml` and updates the index.

### (*ProductType) ModifyProject
Updates an existing project and persists its `project.toml` and the index.

### (*ProjectType) Save
Writes the project data to its `project.toml` file.

### (*ProjectType) Load
Loads the project data from its `project.toml` file.

### (*ProductType) Project
Loads and returns a specific project by ID.

### (*ProductType) LoadProjects
Loads all projects for a product.

### (*Database) LoadAllProjects
Loads all projects for all products in the database.

### (*ProjectType) IngestInputDir
Scans a directory and ingests each regular file as an attachment.

### (*ProjectType) AddAttachmentFromInput
Moves a file from an input folder into the project's attachments and analyses it.

### (*ProjectType) AddAttachmentFromText
Creates an attachment from text content and analyses it.

### (*ProjectType) ActivateRequirementByID
Marks the requirement with the given ID as active.

### (*ProjectType) ActivateRequirementsWhere
Activates all requirements matching the provided predicate.

### (*ProjectType) AddRequirement
Appends a requirement to the project and persists it.

### (*ProjectType) GenerateDesignAspectsAll
Generates design aspects for all requirements and saves the project.

### (*ProjectType) QualityControlScanALL
Runs QualityControlAI for every requirement that has not yet been analysed.

### (*ProjectType) AnalyseAll
Runs QualityControlAI on all non-proposed, non-deleted requirements and saves results.

### (*Requirement) SuggestOthers
Asks the LLM for related requirements and optionally appends them to the project.

### (*Requirement) GenerateDesignAspects
Requests design improvement topics for the requirement and appends them.

### (*DesignAspect) GenerateTemplates
Asks the LLM for requirement templates related to the design aspect and appends them.

### (*ProjectType) ExportExcel
Writes project data to an Excel workbook.

### ImportProjectExcel
Reads an Excel workbook and returns populated `ProjectData`.

### (*ProjectType) ImportExcel
Merges data from an Excel workbook into the project.

### (*ProjectType) Attachments
Returns an `AttachmentManager` for the project.

### (AttachmentManager) AddFromInputFolder
Scans the project's default `input` directory and ingests all files into attachments.

## Package `pmfs`

### NewProject
Ensures the data layout exists, initialises the default LLM client and creates a new project under the first product.

## Package `pmfs/llm`

### NewRateLimitedClient
Wraps a client with simple request-per-second rate limiting.

### SetClient
Replaces the package's LLM client and returns the previous one.

### AnalyzeAttachment
Uploads and analyses a file using the configured client.

### Ask
Sends a prompt to the configured client and returns the response.

### LoadConfig
Loads LLM configuration from `llmconfig.json` or defaults.

### Model
Returns the configured LLM model name.

### RequestsPerSecond
Returns the configured request-per-second limit.

## Package `pmfs/llm/gates`

### Evaluate
Runs the specified quality gates against text using an LLM client.

### GetGate
Retrieves a gate definition by ID.

## Package `pmfs/llm/gemini`

### SetClient
Replaces the Gemini client used by the package and returns the previous one.

### AnalyzeAttachment
Uploads and analyses a file using the Gemini API.

### Ask
Sends a prompt to Gemini and returns the response.

### NewRESTClient
Creates a Gemini REST client configured with an API key and model.

## Package `pmfs/llm/interact`

### RunQuestion
Formats a role-specific question and asks it via the LLM, returning a yes/no result and optional follow-up answer.

## Package `pmfs/llm/prompts`

### RegisterRole
Registers a set of prompts for a role.

### SetTestPrompts
Registers prompts used for the special test role.

### GetPrompts
Returns prompts for a given role.

## Package `pmfs/testgen`

### Verify
Uses Gemini to determine whether code satisfies a specification and returns the verdict.

