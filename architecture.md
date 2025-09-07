\# PMFS Architecture Overview



```mermaid

flowchart TD

  %% --- Attachment ingestion ---
  subgraph Ingestion
    A[Input files] --> B[AttachmentManager.AddFromInputFolder]
    B --> C[Project.AddAttachmentFromInput]
    C --> D[Attachment stored + metadata]
    D --> E[Attachment.Analyze]
    E --> F{LLM.AnalyzeAttachment}
    F --> G[Project.PotentialRequirements]
  end

  %% --- Requirement lifecycle ---
  subgraph "Requirement Workflow"
    G --> H[Project.ActivateRequirement]
    H --> I[Confirmed Requirements]
    I --> J[Requirement.QualityControlAI]
    I --> K[Requirement.GenerateDesignAspects]
    K --> L[Design Aspects + Templates]
    I --> M[Requirement.SuggestOthers]
    M --> G
  end

  %% --- Persistence ---
  G --> N[Project.Save / Database.Save]
  I --> N
