# Architecture diagrams (Mermaid)

These `.mmd` files mirror the diagrams in [architecture.md §14](../architecture.md).

Implementation order and milestones (not a diagram) live in [implementation-plan-deployable-chunks.md](implementation-plan-deployable-chunks.md).

| File | Diagram |
|------|---------|
| [system-context.mmd](system-context.mmd) | Containers: user, orchestration, API, importer, data files |
| [planning-sequence.mmd](planning-sequence.mmd) | Sequence: NL request → validate/rank → response |
| [validation-vs-scoring.mmd](validation-vs-scoring.mmd) | Flow: validation gate vs scoring |

**Export to SVG** (with [Mermaid CLI](https://github.com/mermaid-js/mermaid-cli)):

```bash
npx @mermaid-js/mermaid-cli -i docs/diagrams/system-context.mmd -o docs/diagrams/system-context.svg
```

Repeat per file, or use your editor’s Mermaid preview.
