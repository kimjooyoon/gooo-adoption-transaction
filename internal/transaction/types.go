package transaction

const (
	SourceSchema   = "gooo/adoption-transaction/source/v1"
	ContractSchema = "gooo/adoption-transaction/denominator/v1"
	ManifestSchema = "gooo/adoption-transaction/manifest/v1"
	ReceiptSchema  = "gooo/adoption-transaction/receipt/v1"
	FixedCells     = 16
	FixedCases     = 12
)

type Authority struct {
	RepositoryWrites          int    `json:"repository_writes"`
	LocalTestExecutions       int    `json:"local_test_executions"`
	CrossProjectRequiredGates int    `json:"cross_project_required_gates"`
	ProtectedRepositoryWrites int    `json:"protected_repository_writes"`
	ProductMutationAuthorized bool   `json:"product_mutation_authorized"`
	RootReadmePolicy          string `json:"root_readme_policy"`
}

type BaseTuple struct {
	Base      string `json:"base"`
	Head      string `json:"head"`
	MergeBase string `json:"merge_base"`
}

type Budget struct {
	Limit     int  `json:"limit"`
	Consumed  int  `json:"consumed"`
	Exhausted bool `json:"exhausted"`
}

type Claim struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func (c Claim) HasUnknownTuple() bool {
	return c.State == "UNKNOWN" && c.Stage != "" && c.Step != "" && c.Reason != "" &&
		c.UnknownClass != "" && c.NextOperation != "" && c.BlockedBy != nil && len(c.BlockedBy) > 0
}

type Inventory struct {
	Files              int  `json:"files"`
	Directories        int  `json:"directories"`
	GoFiles            int  `json:"go_files"`
	GoooFiles          int  `json:"gooo_files"`
	PhysicalLines      int  `json:"physical_lines"`
	Artifacts          int  `json:"artifacts"`
	RootReadmeExcluded bool `json:"root_readme_excluded"`
}

type Activity struct {
	Ordinal      int    `json:"ordinal"`
	ID           string `json:"id"`
	SemanticEdge string `json:"semantic_edge"`
	Phase        string `json:"phase"`
}

type SourceDecl struct {
	Schema        string     `json:"schema"`
	Version       string     `json:"version"`
	DenominatorID string     `json:"denominator_id"`
	CellCount     int        `json:"cell_count"`
	Authority     Authority  `json:"authority"`
	Precedence    []string   `json:"precedence"`
	UnknownFields []string   `json:"unknown_fields"`
	Activities    []Activity `json:"activities"`
	SourceDigest  string     `json:"source_digest"`
}

type PhaseCounts struct {
	Prepare       int `json:"PREPARE"`
	Authorize     int `json:"AUTHORIZE"`
	Commit        int `json:"COMMIT"`
	VerifyOrAbort int `json:"VERIFY_OR_ABORT"`
}

type Contract struct {
	Schema     string      `json:"schema"`
	ID         string      `json:"id"`
	Version    string      `json:"version"`
	CellCount  int         `json:"cell_count"`
	Fixed      bool        `json:"fixed"`
	Phases     PhaseCounts `json:"phases"`
	Activities []Activity  `json:"activities"`
}

type FixtureInventory struct {
	Files         int `json:"files"`
	Directories   int `json:"directories"`
	GoFiles       int `json:"go_files"`
	GoooFiles     int `json:"gooo_files"`
	PhysicalLines int `json:"physical_lines"`
}

type Proposal struct {
	ID                      string    `json:"id"`
	Base                    BaseTuple `json:"base"`
	ChangedPaths            []string  `json:"changed_paths"`
	ExpectedPostStateDigest string    `json:"expected_post_state_digest"`
	SourceDigest            string    `json:"source_digest"`
}

type ReleaseIdentity struct {
	TransactionID           string `json:"transaction_id"`
	DenominatorID           string `json:"denominator_id"`
	ProposalDigest          string `json:"proposal_digest"`
	EvaluatorDigest         string `json:"evaluator_digest"`
	ExpectedPostStateDigest string `json:"expected_post_state_digest"`
}

type Attempt struct {
	Ordinal   int    `json:"ordinal"`
	CaseID    string `json:"case_id"`
	Operation string `json:"operation"`
	Outcome   string `json:"outcome"`
	Reason    string `json:"reason"`
}

type Metrics struct {
	WallMS     int `json:"wall_ms"`
	PeakRSSKiB int `json:"peak_rss_kib"`
}

type CaseResult struct {
	ID                      string    `json:"id"`
	Name                    string    `json:"name"`
	State                   string    `json:"state"`
	Decision                string    `json:"decision"`
	Proposer                string    `json:"proposer"`
	Authorizer              string    `json:"authorizer"`
	Base                    BaseTuple `json:"base"`
	AuthorizedChangedPaths  []string  `json:"authorized_changed_paths"`
	ProposedChangedPaths    []string  `json:"proposed_changed_paths"`
	ProposalDigest          string    `json:"proposal_digest"`
	EvaluatorDigest         string    `json:"evaluator_digest"`
	ExpectedPostStateDigest string    `json:"expected_post_state_digest"`
	ObservedPostStateDigest string    `json:"observed_post_state_digest"`
	Budget                  Budget    `json:"budget"`
	CommitAuthority         bool      `json:"commit_authority"`
	ReplayEqual             bool      `json:"replay_equal"`
	Claim                   Claim     `json:"claim"`
	Attempts                []Attempt `json:"attempts"`
	Metrics                 Metrics   `json:"metrics"`
}

type Summary struct {
	Generated int `json:"generated"`
	Closed    int `json:"closed"`
	Unknown   int `json:"unknown"`
	Refuted   int `json:"refuted"`
}

type Manifest struct {
	Schema                   string       `json:"schema"`
	Version                  string       `json:"version"`
	TransactionID            string       `json:"transaction_id"`
	SourceDigest             string       `json:"source_digest"`
	ContractDigest           string       `json:"contract_digest"`
	EvaluatorDigest          string       `json:"evaluator_digest"`
	ProposalDigest           string       `json:"proposal_digest"`
	ReleaseDigest            string       `json:"release_digest"`
	ExpectedPostStateDigest  string       `json:"expected_post_state_digest"`
	Base                     BaseTuple    `json:"base"`
	AuthorizedChangedPaths   []string     `json:"authorized_changed_paths"`
	Precedence               []string     `json:"precedence"`
	UnknownFields            []string     `json:"unknown_fields"`
	Denominator              ContractView `json:"denominator"`
	Budget                   Budget       `json:"budget"`
	Summary                  Summary      `json:"summary"`
	Cases                    []CaseResult `json:"cases"`
	Authority                Authority    `json:"authority"`
	Inventory                Inventory    `json:"inventory"`
	ArtifactNames            []string     `json:"artifact_names"`
	ArtifactCount            int          `json:"artifact_count"`
	AppendOnlyAttempts       []Attempt    `json:"append_only_attempts"`
	ResetOperations          []string     `json:"reset_operations"`
	DeleteOperations         []string     `json:"delete_operations"`
	BudgetExhaustionRecorded bool         `json:"budget_exhaustion_recorded"`
}

type ContractView struct {
	ID         string      `json:"id"`
	CellCount  int         `json:"cell_count"`
	Phases     PhaseCounts `json:"phases"`
	Activities []Activity  `json:"activities"`
}

type Receipt struct {
	Schema                  string    `json:"schema"`
	Kind                    string    `json:"kind"`
	TransactionID           string    `json:"transaction_id"`
	SourceDigest            string    `json:"source_digest"`
	ContractDigest          string    `json:"contract_digest"`
	EvaluatorDigest         string    `json:"evaluator_digest"`
	ProposalDigest          string    `json:"proposal_digest"`
	ReleaseDigest           string    `json:"release_digest"`
	ExpectedPostStateDigest string    `json:"expected_post_state_digest"`
	Base                    BaseTuple `json:"base"`
	AuthorizedChangedPaths  []string  `json:"authorized_changed_paths"`
	CommitAuthority         bool      `json:"commit_authority"`
	AuthorizationGranted    bool      `json:"authorization_granted"`
	PrepareIsAuthority      bool      `json:"prepare_is_authority"`
	Budget                  Budget    `json:"budget"`
	Cases                   []string  `json:"cases"`
	AppendOnlyAttempts      []Attempt `json:"append_only_attempts"`
	ResetOperations         []string  `json:"reset_operations"`
	DeleteOperations        []string  `json:"delete_operations"`
	Metrics                 Metrics   `json:"metrics"`
	Authority               Authority `json:"authority"`
}
