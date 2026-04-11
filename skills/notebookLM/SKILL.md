---
name: notebookLM
description: "Use this skill whenever the user mentions NotebookLM, wants to create YAML for NotebookLM, needs structured slide content, or asks for presentation material generation. Also use when the user wants to convert project documentation into NotebookLM format. Do NOT load for: code implementation, bug fixes, code reviews, deployments, or general slide images (use generate-slide instead). Generates NotebookLM-compatible YAML documents and structured slide content."
allowed-tools: ["Read", "Write", "Edit"]
argument-hint: "[yaml|slides]"
---

# NotebookLM Skill

A collection of skills responsible for document generation.

## Feature Details

| Feature | Details |
|---------|--------|
| **NotebookLM YAML** | See [references/notebooklm-yaml.md](${CLAUDE_SKILL_DIR}/references/notebooklm-yaml.md) |
| **Slide YAML** | See [references/notebooklm-slides.md](${CLAUDE_SKILL_DIR}/references/notebooklm-slides.md) |

## Execution Steps

1. Classify the user's request
2. Read the appropriate reference file from "Feature Details" above
3. Generate according to its contents

---

## 🔧 PDF Page Range Reading (Claude Code 2.1.49+)

A feature for efficiently handling large PDFs.

### Reading with Page Range Specification

```javascript
// Read with page range specification
Read({ file_path: "docs/spec.pdf", pages: "1-10" })

// Check table of contents only
Read({ file_path: "docs/manual.pdf", pages: "1-3" })

// Specific sections only
Read({ file_path: "docs/api-reference.pdf", pages: "25-45" })
```

### Recommended Approaches by Use Case

| Case | Recommended Reading Method | Reason |
|------|---------------------------|--------|
| **PDFs over 100 pages** | Table of contents (1-3) → relevant chapters only | Minimize token consumption |
| **Specification review** | Range specification per section | Read only the necessary parts in detail |
| **API documentation** | Start from endpoint list (table of contents) | Understand overall structure before diving into details |
| **Academic papers** | Abstract + conclusion → body | Grasp key points first |
| **Technical manuals** | Table of contents + troubleshooting chapter | Prioritize practical sections |

### Usage Example for NotebookLM YAML Generation

```markdown
When generating YAML from a large PDF (300-page technical specification):

1. **Read the table of contents** (pages 1-5)
   Read({ file_path: "spec.pdf", pages: "1-5" })
   → Understand the chapter structure

2. **Read the beginning of each chapter** (first 2 pages of each chapter)
   Read({ file_path: "spec.pdf", pages: "10-11" })  // Chapter 1
   Read({ file_path: "spec.pdf", pages: "45-46" })  // Chapter 2
   → Understand each chapter's overview

3. **Read important sections in detail**
   Read({ file_path: "spec.pdf", pages: "78-95" })  // API Reference
   → Extract detailed content

This approach allows efficient YAML generation without reading all 300 pages.
```

### Best Practices

| Principle | Description |
|-----------|-------------|
| **Progressive loading** | Read in order: table of contents → overview → details |
| **Relevant pages only** | Specify only the pages needed for the task |
| **Token conservation** | Reading all pages is a last resort |
| **Structure understanding first** | Understand the big picture from the table of contents before diving into details |

### Comparison with Conventional Methods

| Method | Token Consumption | Processing Time | Accuracy |
|--------|------------------|-----------------|----------|
| **Full page loading** (300 pages) | ~150,000 | Long | High |
| **Page range specification** (30 needed pages) | ~15,000 | Short | High |

→ **90% token reduction and processing time improvement possible**
