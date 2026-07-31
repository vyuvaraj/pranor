# Pranor Project Templates

## Kiro + Pranor (AI-Assisted Development)

To create a new Pranor project with Kiro support:

1. Copy the `kiro-project/` template to your new project directory:
   ```bash
   cp -r templates/kiro-project/ ~/my-new-service/
   cd ~/my-new-service/
   ```

2. Open in Kiro — the steering files tell Kiro how to write Pranor code:
   - `.kiro/steering/pranor.md` — Language syntax & API reference
   - `.kiro/steering/project.md` — Project conventions & patterns

3. Start building:
   ```bash
   pranor run main.pnr --watch
   ```

## What the Steering Files Provide

When Kiro sees these files, it automatically:
- Writes valid `.pnr` syntax (not Go, not TypeScript)
- Uses the correct built-in objects (`log`, `db`, `cache`, `http`, etc.)
- Follows the `?` operator pattern for error handling
- Imports from `stdlib/` correctly
- Runs `pranor build` / `pranor test` / `pranor lint` for verification
- Uses 4-space indentation and `pranor fmt` conventions

## Without Kiro

The template works as a regular Pranor project too — just ignore the `.kiro/` directory.
