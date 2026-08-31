package transaction

import (
	"path/filepath"
	"testing"
)

func TestRunProducesExactAdoptionCorpus(t *testing.T) {
	root := filepath.Join("..", "..")
	manifest, err := Run(
		filepath.Join(root, "examples", "adoption-transaction-v1", "transaction.gooo"),
		filepath.Join(root, "contracts", "denominator-v1.json"),
		t.TempDir(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateManifest(manifest); err != nil {
		t.Fatal(err)
	}
	if !manifest.Cases[0].CommitAuthority || manifest.Cases[0].Decision != "COMMIT_CLOSED" {
		t.Fatalf("valid commit did not close: %+v", manifest.Cases[0])
	}
	for _, item := range manifest.Cases {
		if item.State == "UNKNOWN" && !item.Claim.HasUnknownTuple() {
			t.Fatalf("unknown claim is incomplete: %+v", item.Claim)
		}
	}
}
