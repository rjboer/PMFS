\# PMFS Architecture Overview



```mermaid

flowchart TD

&nbsp;   %% --- Attachment ingestion ---

&nbsp;   subgraph Ingestion

&nbsp;       A\[Input files] --> B\[AttachmentManager.AddFromInputFolder]

&nbsp;       B --> C\[Project.AddAttachmentFromInput]

&nbsp;       C --> D\[Attachment stored + metadata]

&nbsp;       D --> E\[Attachment.Analyze]

&nbsp;       E --> F{LLM.AnalyzeAttachment}

&nbsp;       F --> G\[Project.PotentialRequirements]

&nbsp;   end



&nbsp;   %% --- Requirement lifecycle ---

&nbsp;   subgraph Requirement Workflow

&nbsp;       G --> H\[Project.ActivateRequirement]

&nbsp;       H --> I\[Confirmed Requirements]

&nbsp;       I --> J\[Requirement.QualityControlAI]

&nbsp;       I --> K\[Requirement.GenerateDesignAspects]

&nbsp;       K --> L\[Design Aspects + Templates]

&nbsp;       I --> M\[Requirement.SuggestOthers]

&nbsp;       M --> G

&nbsp;   end



&nbsp;   %% --- Persistence ---

&nbsp;   G \& I --> N\[Project.Save / Database.Save]



```

