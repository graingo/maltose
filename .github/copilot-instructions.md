# Copilot review instructions for graingo/maltose

When performing pull request reviews in this repository:

- Always respond in Simplified Chinese.
- Keep review comments concise, actionable, and technically specific.
- Prioritize correctness, safety, maintainability, and testability.
- Focus on configuration correctness, error handling, concurrency safety, CI/CD impact, and backward compatibility.
- For Go code, pay special attention to:
  - nil handling
  - error wrapping and propagation
  - goroutine leaks
  - channel misuse
  - race conditions
  - config parsing and default values
- For workflow or release related changes, pay special attention to:
  - secret usage
  - permissions scope
  - branch trigger conditions
  - accidental breaking of existing release/test flows
- Avoid generic praise unless it helps explain why a change is safe.
- If no significant issue is found, provide a short Chinese summary and light-risk observations instead of forcing problems.
- Prefer bullet points and short sections.
