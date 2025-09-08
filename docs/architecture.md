# PMFS Architecture Overview

PMFS maintains a single collection of requirements. Each requirement carries
condition flags that describe its lifecycle state. The diagram below illustrates
how artifacts are ingested and how requirements move through this lifecycle
using these flags.

```mermaid
flowchart TD

  LLM[(LLM Service)]
  GATE[(Gates)]

  subgraph Ingestion
    A[Add files to Input folder]
    A --> C[Project.AddAttachmentFromInput]
    C --> D[Attachment stored in folder + metadata]
    D --> E[Attachment.Analyze]
    E --> LLM
    LLM --> X1[Create Intelligence & Design Aspects]
    X1 --> X2[Intelligence]
    X1 --> X3[Design Aspects]
    LLM --> RP[Requirement; Proposed, AIgenerated]
    X2 --> RP
    X3 --> X4[Generate requirements based on design aspects]
    X4 --> RP
  end

  subgraph "Requirement Lifecycle"
    RP --> H[Project.ActivateRequirementByID]
    H --> RA[Requirement, Active]
    RA --> J[Requirement.QualityControlAI]
    J --> LLM
    J --> GATE
    LLM --> J
    GATE --> J
    J --> K[Gate Results]
    K --> RA
    RA --> L[Requirement.GenerateDesignAspects]
    L --> N[Design Aspects]
    N --> O[Generate requirements based on design aspects]
    O --> RP
    RA --> M[Requirement.SuggestOthers]
    M --> RP
    RA --> RD[Mark requirement Deleted]
    RD --> RP
    RP --> P[Deduplication of requirements]
  end

  RP --> Q[Project.Save / Database.Save]
  RA --> Q

  Q --> R1[Project.ExportExcel]
  R1 --> S[Excel Workbook]
  S --> T[ImportProjectExcel]
  T --> Q

```

## Requirement Condition Flags

- **Proposed** – candidate requirement awaiting activation. Proposed
  requirements are skipped by analysis and gating.
- **AIgenerated** – indicates the requirement originated from the LLM. User-created
  requirements leave this false.
- **AIanalyzed** – requirement has already been processed by the LLM and is
  skipped during subsequent analyses.
- **Active** – requirement is approved and participates in analysis, gating, and
  export.
- **Deleted** – requirement has been removed from active consideration but is
  retained for history and ignored by processing.

### Typical transitions

1. `Proposed` (often `AIgenerated`) → `Active` via activation.
2. `Proposed` → `Deleted` when a candidate is discarded.
3. `Active` → `Deleted` if an accepted requirement is later removed.

