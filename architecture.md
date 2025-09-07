\# PMFS Architecture Overview



```mermaid

flowchart TD

  %% --- Attachment ingestion ---
  subgraph Ingestion


    A[Add files to Input folder]
    A --> C[Project.AddAttachmentFromInput]
    C --> D[Attachment stored in folder + metadata]
    D --> E[Attachment.Analyze]
    E --> F{LLM.AnalyzeAttachment}
    F --> G[Project.PotentialRequirements]
  end

  %% --- Requirement lifecycle ---
  subgraph "Requirement Workflow, assume we have a list of requirements"
    G --> H[Project.ActivateRequirement]
    H --> I[list of Requirements]
    I --> J[Requirement.QualityControlAI]
    I --> K[Requirement.GenerateDesignAspects]
    K --> L[Design Aspects ]
    L --> N[Generate requirements based on design aspects]
    N --> G
    I --> M[Requirement.SuggestOthers]
    M --> G
    G --> P[Deduplication of requirements]
    H --> P

  end

  %% --- Persistence ---
  G --> O[Project.Save / Database.Save]
  I --> O
