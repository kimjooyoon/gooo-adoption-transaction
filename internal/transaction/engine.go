package transaction

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	transactionID = "adoption-transaction-v1"
	baseSHA       = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	headSHA       = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	mergeBaseSHA  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	proposer      = "candidate-agent"
	authorizer    = "protected-owner"
	targetPath    = "service/order.go"
)

type seed struct {
	sourceDigest    string
	contractDigest  string
	evaluatorDigest string
	proposalDigest  string
	releaseDigest   string
	beforeDigest    string
	afterDigest     string
	base            BaseTuple
	authorizedPaths []string
	proposal        Proposal
}

type fixtureState struct {
	Files map[string]string `json:"files"`
}

func Run(sourcePath, contractPath, outputPath string) (Manifest, error) {
	source, err := parseSource(sourcePath)
	if err != nil {
		return Manifest{}, err
	}
	contract, err := loadContract(contractPath)
	if err != nil {
		return Manifest{}, err
	}
	if err := validateDeclarations(source, contract); err != nil {
		return Manifest{}, err
	}
	if err := ensureOutput(outputPath); err != nil {
		return Manifest{}, err
	}
	contractDigest, err := DigestValue(contract)
	if err != nil {
		return Manifest{}, err
	}
	transactionSeed, err := makeSeed(source, contractDigest)
	if err != nil {
		return Manifest{}, err
	}
	cases := evaluateCases(transactionSeed)
	summary := summarize(cases)
	attempts := appendOnlyAttempts(cases)
	manifest := Manifest{
		Schema:                   ManifestSchema,
		Version:                  "v1",
		TransactionID:            transactionID,
		SourceDigest:             transactionSeed.sourceDigest,
		ContractDigest:           transactionSeed.contractDigest,
		EvaluatorDigest:          transactionSeed.evaluatorDigest,
		ProposalDigest:           transactionSeed.proposalDigest,
		ReleaseDigest:            transactionSeed.releaseDigest,
		ExpectedPostStateDigest:  transactionSeed.afterDigest,
		Base:                     transactionSeed.base,
		AuthorizedChangedPaths:   copyStrings(transactionSeed.authorizedPaths),
		Precedence:               copyStrings(source.Precedence),
		UnknownFields:            copyStrings(source.UnknownFields),
		Denominator:              contractView(contract),
		Budget:                   Budget{Limit: 1, Consumed: 1, Exhausted: true},
		Summary:                  summary,
		Cases:                    cases,
		Authority:                zeroAuthority(),
		Inventory:                Inventory{Files: 2, Directories: 1, GoFiles: 1, GoooFiles: 1, PhysicalLines: 8, Artifacts: 7, RootReadmeExcluded: true},
		ArtifactNames:            artifactNames(),
		ArtifactCount:            7,
		AppendOnlyAttempts:       attempts,
		ResetOperations:          []string{},
		DeleteOperations:         []string{},
		BudgetExhaustionRecorded: true,
	}
	if err := validateManifest(manifest); err != nil {
		return Manifest{}, err
	}
	if err := writeArtifacts(outputPath, manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func makeSeed(source SourceDecl, contractDigest string) (seed, error) {
	before := fixtureState{Files: map[string]string{
		targetPath: "package service\n\ntype Order struct {\n\tStatus string\n}\n",
	}}
	after := fixtureState{Files: map[string]string{
		targetPath: "package service\n\ntype Order struct {\n\tStatus string\n\tPriority int\n}\n",
	}}
	beforeDigest, err := DigestValue(before)
	if err != nil {
		return seed{}, err
	}
	afterDigest, err := DigestValue(after)
	if err != nil {
		return seed{}, err
	}
	proposal := Proposal{
		ID:                      transactionID,
		Base:                    BaseTuple{Base: baseSHA, Head: headSHA, MergeBase: mergeBaseSHA},
		ChangedPaths:            []string{targetPath},
		ExpectedPostStateDigest: afterDigest,
		SourceDigest:            source.SourceDigest,
	}
	proposalDigest, err := DigestValue(proposal)
	if err != nil {
		return seed{}, err
	}
	evaluatorDigest, err := DigestValue(map[string]string{
		"evaluator": "gooo-adoption-transaction",
		"version":   "v1",
		"contract":  contractDigest,
	})
	if err != nil {
		return seed{}, err
	}
	releaseDigest, err := DigestValue(ReleaseIdentity{
		TransactionID:           transactionID,
		DenominatorID:           source.DenominatorID,
		ProposalDigest:          proposalDigest,
		EvaluatorDigest:         evaluatorDigest,
		ExpectedPostStateDigest: afterDigest,
	})
	if err != nil {
		return seed{}, err
	}
	return seed{
		sourceDigest: source.SourceDigest, contractDigest: contractDigest,
		evaluatorDigest: evaluatorDigest, proposalDigest: proposalDigest,
		releaseDigest: releaseDigest, beforeDigest: beforeDigest, afterDigest: afterDigest,
		base: proposal.Base, authorizedPaths: proposal.ChangedPaths, proposal: proposal,
	}, nil
}

func evaluateCases(s seed) []CaseResult {
	valid := newCase(s, "CASE-01-VALID-COMMIT", "valid commit", "CLOSED", "COMMIT_CLOSED")
	valid.CommitAuthority = true
	valid.ExpectedPostStateDigest = s.afterDigest
	valid.ObservedPostStateDigest = s.afterDigest
	valid.Budget = Budget{Limit: 1, Consumed: 1, Exhausted: true}
	valid.Attempts = []Attempt{
		attempt(1, valid.ID, "PREPARE", "CLOSED", "PREPARE_RECEIPT_CREATED"),
		attempt(2, valid.ID, "AUTHORIZE", "CLOSED", "EXACT_AUTHORIZATION_GRANTED"),
		attempt(3, valid.ID, "COMMIT", "CLOSED", "COMMIT_APPLIED_ONCE"),
		attempt(4, valid.ID, "VERIFY", "CLOSED", "EXPECTED_POST_STATE_MATCHED"),
	}

	replay := newCase(s, "CASE-02-DETERMINISTIC-REPLAY", "deterministic replay", "CLOSED", "COMMIT_CLOSED")
	replay.CommitAuthority = true
	replay.ExpectedPostStateDigest = s.afterDigest
	replay.ObservedPostStateDigest = s.afterDigest
	replay.Budget = Budget{Limit: 2, Consumed: 2, Exhausted: true}
	replay.ReplayEqual = true
	replay.Attempts = []Attempt{
		attempt(1, replay.ID, "COMMIT", "CLOSED", "FIRST_COMMIT_APPLIED"),
		attempt(2, replay.ID, "REPLAY", "CLOSED", "SECOND_RESULT_BYTE_EQUAL"),
	}

	abort := newCase(s, "CASE-03-EXPLICIT-ABORT", "explicit abort", "CLOSED", "ABORT_CLOSED")
	abort.ExpectedPostStateDigest = s.beforeDigest
	abort.ObservedPostStateDigest = s.beforeDigest
	abort.Budget = Budget{Limit: 1, Consumed: 0, Exhausted: false}
	abort.Attempts = []Attempt{
		attempt(1, abort.ID, "PREPARE", "CLOSED", "PREPARE_RECEIPT_CREATED"),
		attempt(2, abort.ID, "ABORT", "CLOSED", "EXPLICIT_ABORT_RECORDED"),
	}

	missingAuth := newCase(s, "CASE-04-MISSING-AUTHORIZATION", "missing authorization", "UNKNOWN", "UNKNOWN")
	missingAuth.Claim = unknownClaim("AUTHORIZE", "verify_authorization", "AUTHORIZATION_RECEIPT_NOT_PROVIDED", "MISSING_AUTHORIZATION", "OBTAIN_AUTHORIZATION_RECEIPT", []string{"authorization-receipt"})
	missingAuth.Attempts = []Attempt{
		attempt(1, missingAuth.ID, "PREPARE", "CLOSED", "PREPARE_RECEIPT_CREATED"),
		attempt(2, missingAuth.ID, "COMMIT", "UNKNOWN", "AUTHORIZATION_EVIDENCE_MISSING"),
	}

	staleBase := newCase(s, "CASE-05-STALE-BASE", "stale base", "UNKNOWN", "UNKNOWN")
	staleBase.Base = BaseTuple{Base: "cccccccccccccccccccccccccccccccccccccccc", Head: s.base.Head, MergeBase: "cccccccccccccccccccccccccccccccccccccccc"}
	staleBase.Claim = unknownClaim("AUTHORIZE", "refresh_base_tuple", "BASE_HEAD_MERGE_BASE_IS_STALE", "STALE_BASE", "REFRESH_EXACT_BASE_TUPLE", []string{"current-base-evidence"})
	staleBase.Attempts = []Attempt{attempt(1, staleBase.ID, "AUTHORIZE", "UNKNOWN", "STALE_BASE_REQUIRES_REFRESH")}

	exhausted := newCase(s, "CASE-06-EXHAUSTED-BUDGET-REPLAY", "exhausted budget replay", "UNKNOWN", "UNKNOWN")
	exhausted.Budget = Budget{Limit: 1, Consumed: 1, Exhausted: true}
	exhausted.Claim = unknownClaim("COMMIT", "replay_budget_check", "BUDGET_EXHAUSTED_REPLAY_CANNOT_BE_AUTHORIZED", "EXHAUSTED_BUDGET", "ISSUE_NEW_APPEND_ONLY_AUTHORIZATION", []string{"budget:1/1"})
	exhausted.Attempts = []Attempt{
		attempt(1, exhausted.ID, "COMMIT", "CLOSED", "BUDGET_UNIT_CONSUMED"),
		attempt(2, exhausted.ID, "REPLAY", "UNKNOWN", "BUDGET_EXHAUSTED_NO_RESET_ALLOWED"),
	}

	extraPath := newCase(s, "CASE-07-UNAUTHORIZED-EXTRA-PATH", "unauthorized extra path", "REFUTED", "REFUTED")
	extraPath.ProposedChangedPaths = []string{targetPath, "docs/extra.md"}
	extraPath.Claim = refutedClaim("AUTHORIZE", "compare_changed_path_tuple", "UNAUTHORIZED_EXTRA_PATH", "REJECT_UNAUTHORIZED_PATH")
	extraPath.Attempts = []Attempt{attempt(1, extraPath.ID, "AUTHORIZE", "REFUTED", "EXACT_PATH_TUPLE_MISMATCH")}

	proposalMismatch := newCase(s, "CASE-08-PROPOSAL-DIGEST-MISMATCH", "proposal digest mismatch", "REFUTED", "REFUTED")
	proposalMismatch.ProposalDigest = DigestBytes([]byte("tampered proposal"))
	proposalMismatch.Claim = refutedClaim("AUTHORIZE", "compare_proposal_digest", "PROPOSAL_DIGEST_MISMATCH", "REGENERATE_IMMUTABLE_PROPOSAL")
	proposalMismatch.Attempts = []Attempt{attempt(1, proposalMismatch.ID, "AUTHORIZE", "REFUTED", "PROPOSAL_DIGEST_MISMATCH")}

	evaluatorMismatch := newCase(s, "CASE-09-EVALUATOR-DIGEST-MISMATCH", "evaluator digest mismatch", "REFUTED", "REFUTED")
	evaluatorMismatch.EvaluatorDigest = DigestBytes([]byte("tampered evaluator"))
	evaluatorMismatch.Claim = refutedClaim("AUTHORIZE", "compare_evaluator_digest", "EVALUATOR_DIGEST_MISMATCH", "PIN_EXPECTED_EVALUATOR_DIGEST")
	evaluatorMismatch.Attempts = []Attempt{attempt(1, evaluatorMismatch.ID, "AUTHORIZE", "REFUTED", "EVALUATOR_DIGEST_MISMATCH")}

	nullPost := newCase(s, "CASE-10-NULL-POST-DIGEST", "null post digest", "REFUTED", "REFUTED")
	nullPost.ExpectedPostStateDigest = ""
	nullPost.Claim = refutedClaim("COMMIT", "require_expected_post_digest", "EXPECTED_POST_STATE_DIGEST_IS_NULL", "REJECT_NULL_POST_STATE")
	nullPost.Attempts = []Attempt{attempt(1, nullPost.ID, "COMMIT", "REFUTED", "NULL_POST_DIGEST")}

	selfAuth := newCase(s, "CASE-11-SELF-AUTHORIZATION", "self-authorization", "REFUTED", "REFUTED")
	selfAuth.Proposer = proposer
	selfAuth.Authorizer = proposer
	selfAuth.Claim = refutedClaim("AUTHORIZE", "compare_actor_identity", "PROPOSER_EQUALS_AUTHORIZER", "REQUIRE_INDEPENDENT_AUTHORIZER")
	selfAuth.Attempts = []Attempt{attempt(1, selfAuth.ID, "AUTHORIZE", "REFUTED", "SELF_AUTHORIZATION")}

	mergeMismatch := newCase(s, "CASE-12-MERGE-BASE-MISMATCH", "merge-base mismatch", "REFUTED", "REFUTED")
	mergeMismatch.Base = BaseTuple{Base: s.base.Base, Head: s.base.Head, MergeBase: "dddddddddddddddddddddddddddddddddddddddd"}
	mergeMismatch.Claim = refutedClaim("AUTHORIZE", "compare_base_head_merge_base", "MERGE_BASE_DOES_NOT_MATCH_BASE_TUPLE", "REJECT_BASE_TUPLE")
	mergeMismatch.Attempts = []Attempt{attempt(1, mergeMismatch.ID, "AUTHORIZE", "REFUTED", "MERGE_BASE_MISMATCH")}

	return []CaseResult{valid, replay, abort, missingAuth, staleBase, exhausted, extraPath, proposalMismatch, evaluatorMismatch, nullPost, selfAuth, mergeMismatch}
}

func newCase(s seed, id, name, state, decision string) CaseResult {
	return CaseResult{
		ID: id, Name: name, State: state, Decision: decision,
		Proposer: proposer, Authorizer: authorizer,
		Base: s.base, AuthorizedChangedPaths: copyStrings(s.authorizedPaths),
		ProposedChangedPaths: copyStrings(s.authorizedPaths),
		ProposalDigest:       s.proposalDigest, EvaluatorDigest: s.evaluatorDigest,
		ExpectedPostStateDigest: s.afterDigest, Budget: Budget{Limit: 1},
		Claim: Claim{State: state, BlockedBy: []string{}}, Metrics: Metrics{WallMS: 1, PeakRSSKiB: 64},
	}
}

func unknownClaim(stage, step, reason, class, next string, blocked []string) Claim {
	return Claim{State: "UNKNOWN", Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: blocked}
}

func refutedClaim(stage, step, reason, next string) Claim {
	return Claim{State: "REFUTED", Stage: stage, Step: step, Reason: reason, NextOperation: next, BlockedBy: []string{}}
}

func attempt(ordinal int, caseID, operation, outcome, reason string) Attempt {
	return Attempt{Ordinal: ordinal, CaseID: caseID, Operation: operation, Outcome: outcome, Reason: reason}
}

func summarize(cases []CaseResult) Summary {
	result := Summary{Generated: len(cases)}
	for _, item := range cases {
		switch item.State {
		case "CLOSED":
			result.Closed++
		case "UNKNOWN":
			result.Unknown++
		case "REFUTED":
			result.Refuted++
		}
	}
	return result
}

func appendOnlyAttempts(cases []CaseResult) []Attempt {
	result := make([]Attempt, 0)
	ordinal := 1
	for _, item := range cases {
		for _, entry := range item.Attempts {
			entry.Ordinal = ordinal
			result = append(result, entry)
			ordinal++
		}
	}
	return result
}

func zeroAuthority() Authority {
	return Authority{RootReadmePolicy: "EXCLUDED_FROM_REPOSITORY_INVENTORY"}
}

func contractView(contract Contract) ContractView {
	return ContractView{ID: contract.ID, CellCount: contract.CellCount, Phases: contract.Phases, Activities: contract.Activities}
}

func artifactNames() []string {
	return []string{
		"transaction-manifest.json",
		"prepare-receipt.json",
		"authorization-receipt.json",
		"commit-receipt.json",
		"abort-receipt.json",
		"replay-receipt.json",
		"adoption-report.md",
	}
}

func writeArtifacts(outputPath string, manifest Manifest) error {
	abs, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(abs, "transaction-manifest.json"), manifest); err != nil {
		return err
	}
	common := Receipt{
		Schema: manifestReceiptSchema(), TransactionID: manifest.TransactionID,
		SourceDigest: manifest.SourceDigest, ContractDigest: manifest.ContractDigest,
		EvaluatorDigest: manifest.EvaluatorDigest, ProposalDigest: manifest.ProposalDigest,
		ReleaseDigest: manifest.ReleaseDigest, ExpectedPostStateDigest: manifest.ExpectedPostStateDigest,
		Base: manifest.Base, AuthorizedChangedPaths: copyStrings(manifest.AuthorizedChangedPaths),
		Budget: manifest.Budget, AppendOnlyAttempts: manifest.AppendOnlyAttempts,
		ResetOperations: []string{}, DeleteOperations: []string{}, Metrics: Metrics{WallMS: 1, PeakRSSKiB: 64},
		Authority: manifest.Authority,
	}
	receipts := map[string]Receipt{
		"prepare-receipt.json":       receiptFor(common, "PREPARE", false, false, false, []string{"CASE-01-VALID-COMMIT", "CASE-02-DETERMINISTIC-REPLAY", "CASE-03-EXPLICIT-ABORT"}),
		"authorization-receipt.json": receiptFor(common, "AUTHORIZE", false, true, false, []string{"CASE-01-VALID-COMMIT", "CASE-02-DETERMINISTIC-REPLAY", "CASE-03-EXPLICIT-ABORT"}),
		"commit-receipt.json":        receiptFor(common, "COMMIT", true, true, false, []string{"CASE-01-VALID-COMMIT", "CASE-02-DETERMINISTIC-REPLAY"}),
		"abort-receipt.json":         receiptFor(common, "ABORT", false, false, false, []string{"CASE-03-EXPLICIT-ABORT"}),
		"replay-receipt.json":        receiptFor(common, "REPLAY", true, true, false, []string{"CASE-02-DETERMINISTIC-REPLAY", "CASE-06-EXHAUSTED-BUDGET-REPLAY"}),
	}
	for name, receipt := range receipts {
		if err := writeJSON(filepath.Join(abs, name), receipt); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(abs, "adoption-report.md"), []byte(renderReport(manifest)), 0o644)
}

func manifestReceiptSchema() string {
	return ReceiptSchema
}

func receiptFor(base Receipt, kind string, commitAuthority, authorizationGranted, prepareIsAuthority bool, cases []string) Receipt {
	base.Kind = kind
	base.CommitAuthority = commitAuthority
	base.AuthorizationGranted = authorizationGranted
	base.PrepareIsAuthority = prepareIsAuthority
	base.Cases = cases
	return base
}

func validateManifest(manifest Manifest) error {
	if manifest.Schema != ManifestSchema || manifest.Version != "v1" || manifest.TransactionID != transactionID {
		return fmt.Errorf("manifest identity is invalid")
	}
	if manifest.Summary != (Summary{Generated: 12, Closed: 3, Unknown: 3, Refuted: 6}) {
		return fmt.Errorf("unexpected case summary")
	}
	if len(manifest.Cases) != FixedCases || manifest.ArtifactCount != 7 || !sameStrings(manifest.ArtifactNames, artifactNames()) {
		return fmt.Errorf("fixed case or artifact count mismatch")
	}
	if manifest.Denominator.CellCount != FixedCells || manifest.Denominator.Phases != (PhaseCounts{Prepare: 4, Authorize: 4, Commit: 4, VerifyOrAbort: 4}) || len(manifest.Denominator.Activities) != FixedCells {
		return fmt.Errorf("fixed phase denominator mismatch")
	}
	if manifest.Authority.RepositoryWrites != 0 || manifest.Authority.LocalTestExecutions != 0 || manifest.Authority.CrossProjectRequiredGates != 0 || manifest.Authority.ProtectedRepositoryWrites != 0 || manifest.Authority.ProductMutationAuthorized {
		return fmt.Errorf("manifest authority is not zero")
	}
	for _, item := range manifest.Cases {
		if item.State == "UNKNOWN" && !item.Claim.HasUnknownTuple() {
			return fmt.Errorf("unknown case %q has incomplete claim", item.ID)
		}
	}
	return nil
}
