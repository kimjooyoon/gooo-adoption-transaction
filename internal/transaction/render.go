package transaction

import (
	"fmt"
	"strings"
)

func renderReport(manifest Manifest) string {
	var builder strings.Builder
	builder.WriteString("# Adoption transaction report\n\n")
	builder.WriteString("Resolution precedence: `REFUTED > UNKNOWN > CLOSED`\n\n")
	builder.WriteString("## Exact summary\n\n")
	builder.WriteString("| generated | closed | unknown | refuted |\n")
	builder.WriteString("|---:|---:|---:|---:|\n")
	fmt.Fprintf(&builder, "| %d | %d | %d | %d |\n\n", manifest.Summary.Generated, manifest.Summary.Closed, manifest.Summary.Unknown, manifest.Summary.Refuted)
	builder.WriteString("## Transaction identity\n\n")
	fmt.Fprintf(&builder, "- transaction: `%s`\n- source digest: `%s`\n- contract digest: `%s`\n- evaluator digest: `%s`\n- proposal digest: `%s`\n- release digest: `%s`\n- base/head/merge-base: `%s` / `%s` / `%s`\n- exact authorized paths: `%s`\n- expected post-state digest: `%s`\n\n", manifest.TransactionID, manifest.SourceDigest, manifest.ContractDigest, manifest.EvaluatorDigest, manifest.ProposalDigest, manifest.ReleaseDigest, manifest.Base.Base, manifest.Base.Head, manifest.Base.MergeBase, strings.Join(manifest.AuthorizedChangedPaths, ","), manifest.ExpectedPostStateDigest)
	builder.WriteString("## Cases\n\n")
	builder.WriteString("| id | state | decision | changed paths | expected/observed post digest | claim reason |\n")
	builder.WriteString("|---|---|---|---|---|---|\n")
	for _, item := range manifest.Cases {
		fmt.Fprintf(&builder, "| `%s` | `%s` | `%s` | `%s` | `%s` / `%s` | `%s` |\n", item.ID, item.State, item.Decision, strings.Join(item.ProposedChangedPaths, ","), item.ExpectedPostStateDigest, item.ObservedPostStateDigest, item.Claim.Reason)
	}
	builder.WriteString("\n## Authority and evidence\n\n")
	fmt.Fprintf(&builder, "- repository writes: `%d`\n- local test executions: `%d`\n- cross-project required gates: `%d`\n- protected repository writes: `%d`\n- artifact count: `%d`\n- fixture files/directories: `%d`/%d\n- Go/Gooo files: `%d`/%d\n- physical lines: `%d`\n- wall_ms: `%d`\n- peak_rss_kib: `%d`\n- root README: excluded\n- reset operations: none\n- delete operations: none\n", manifest.Authority.RepositoryWrites, manifest.Authority.LocalTestExecutions, manifest.Authority.CrossProjectRequiredGates, manifest.Authority.ProtectedRepositoryWrites, manifest.ArtifactCount, manifest.Inventory.Files, manifest.Inventory.Directories, manifest.Inventory.GoFiles, manifest.Inventory.GoooFiles, manifest.Inventory.PhysicalLines, manifest.Cases[0].Metrics.WallMS, manifest.Cases[0].Metrics.PeakRSSKiB)
	return builder.String()
}
