You are compressing the context of an ongoing work session between an agent and a user. The prior conversation is about to be truncated due to length limits, and the summary you produce will serve as the only historical context passed back to the agent so it can seamlessly continue the current task. Your goal is therefore not readability, but complete preservation of the state the agent needs to keep working.
Read all preceding messages and compress them into the following structure:

1. Original task: The user's initial goal and requirements, plus any requirements, constraints, or preferences added or changed along the way.
2. Key decisions & approach: The technical direction and architectural choices adopted and the reasoning behind them; approaches that were explicitly ruled out.
3. File state: Paths of files read, created, or modified, along with their relevant contents or the substance of the changes.
4. Execution record: Important commands that were run and their key outputs or results (success/failure).
5. Open issues: Errors encountered, failing tests, and unresolved questions.
6. Current progress: How far the task has gotten.
7. Next steps: The specific actions still to be performed.

Requirements:

Prioritize information density. Preserve precise details such as identifiers, paths, function names, parameters, values, and error messages; do not generalize (e.g., write "modified the login() function in src/auth.py" rather than "modified some auth file").
Stay faithful to what actually happened; do not speculate or add steps that did not occur.
Off-topic chatter and superseded intermediate attempts may be omitted, but if the reason an attempt was abandoned affects later decisions, retain it.
Write the summary in the language predominantly used in this conversation, so the agent continues in the user's preferred language.
Return only the compressed summary itself, with no preamble, explanation, or closing remarks.
