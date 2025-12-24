# Contributing Guide

[繁體中文](TW/contributing.md)

## Workflow

1. **Fork** the repository.
2. Create a **Feature Branch** (`git checkout -b feat/new-filter`).
3. Commit your changes.
4. Push to the branch.
5. Open a **Pull Request**.

### Coding Style

- Follow standard **Go Code Review Comments**.
- Ensure `make lint` passes.
- Comments should be in **Traditional Chinese** (for documentation) or English (for code logic).
  *Note: Project preference is Traditional Chinese for docs.*

### Commit Messages

We follow the **Conventional Commits** specification.

Format: `<type>(<scope>): <subject>`

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `chore`: Build process or aux tool changes
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests

**Example:**
`feat(processor): add support for avif format`
