# Plan Review Flow

## Steps

1. Read `Plans.md`
2. Review from 5 perspectives:
   - **Clarity**: Are task descriptions clear?
   - **Feasibility**: Is it technically feasible?
   - **Dependencies**: Are inter-task dependencies correct? (Depends column entries match actual dependencies?)
   - **Acceptance**: Are completion criteria (DoD column) defined and verifiable?
   - **Value**: Does this task solve a user problem? Is "whose problem" stated? Were alternatives considered? Are there Elephants (problems everyone notices but no one addresses)?
3. DoD / Depends column quality checks:
   - Tasks with empty DoD → Warning ("Completion criteria undefined")
   - Unverifiable DoD ("looks good", "works properly") → Warning + concretization suggestion
   - Depends referencing non-existent task numbers → Error
   - Circular dependencies → Error
4. Present improvement suggestions
