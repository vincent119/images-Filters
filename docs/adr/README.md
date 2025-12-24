# Architecture Decision Records (ADR)

[繁體中文](../../TW/adr/README.md)

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

### Template

When adding a new ADR, please use the following template (file name format: `NNNN-title-in-kebab-case.md`):

```markdown
# NNNN. Title of the Decision

- **Status**: [Proposed | Accepted | Superseded | Deprecated]
- **Date**: YYYY-MM-DD
- **Authors**: [Name]

## Context
Describe the problem or opportunity that motivates this decision. Explain the forces at play (technological, political, social, project local).

## Decision
Explain the decision being made. Use active voice ("We will...").

## Consequences
Describe the resulting context after applying this decision. All consequences should be listed here, not just the positive ones.
- **Positive**: ...
- **Negative**: ...
- **Risks**: ...
```

### Index

| ID   | Title                         | Status   | Date       |
|------|-------------------------------|----------|------------|
| 0001 | Record Architecture Decisions | Accepted | 2024-03-20 |
