# Gooo Adoption Transaction

Gooo Adoption Transaction is a small executable conformance laboratory for
two-phase adoption. It simulates a candidate change only inside a
caller-owned temporary fixture and requires:

```text
PREPARE → AUTHORIZE → COMMIT → VERIFY / ABORT
```

The released `.gooo` declaration has a fixed 16-cell denominator: four cells
for each phase, with one-to-one bindings to sixteen released meta activities.
The executable corpus has exactly twelve cases: 3 `CLOSED`, 3 `UNKNOWN`, and
6 `REFUTED`.

`gooo-adoption-transaction run` writes exactly seven artifacts under the
caller-provided empty temporary directory:

```text
transaction-manifest.json
prepare-receipt.json
authorization-receipt.json
commit-receipt.json
abort-receipt.json
replay-receipt.json
adoption-report.md
```

The evaluator accepts a commit only when the immutable proposal and evaluator
digests, exact base/head/merge-base tuple, exact authorized changed-path tuple,
budget receipt, and expected post-state digest all match. A prepare receipt is
not commit authority. Resolution precedence is `REFUTED > UNKNOWN > CLOSED`.
Unknown results preserve `stage`, `step`, `reason`, `unknown_class`,
`next_operation`, and `blocked_by`.

The lab records repository writes, local test executions, cross-project gates,
peak RSS, wall time, file/directory/Go/Gooo/physical-line counts, artifact
count, and append-only failed attempts. It never writes the protected input
repository.

## Validation

GitHub Actions is the validation authority. The workflow uses Go 1.27 for
formatting, vet, tests, build, and conformance. Generated outputs are written
only beneath the runner's temporary directory.
