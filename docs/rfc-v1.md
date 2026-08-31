# Adoption transaction v1

Adoption is a two-phase protocol. PREPARE creates immutable evidence but never
grants commit authority. AUTHORIZE binds the proposal, evaluator, exact Git
base tuple, exact changed-path tuple, actor identity, and budget. COMMIT may
apply only the authorized tuple to a caller-owned temporary fixture. VERIFY
closes a commit only after the exact expected post-state digest matches;
otherwise the transaction explicitly ABORTs or remains UNKNOWN.

The fixed denominator is sixteen cells, four per phase, and the released
`.gooo` activities bind one-to-one to the contract. The fixed executable
corpus has twelve cases with 3 CLOSED, 3 UNKNOWN, and 6 REFUTED outcomes.

Unknown claims always carry six fields: `stage`, `step`, `reason`,
`unknown_class`, `next_operation`, and `blocked_by`. Semantic contradictions
are evaluated first, so `REFUTED > UNKNOWN > CLOSED`.

All failed attempts and budget exhaustion are append-only receipt entries.
There is no reset or delete operation. Repository writes, local test
executions, and cross-project required gates remain zero.
