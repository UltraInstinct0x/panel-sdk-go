package panelsdk

import "encoding/json"

type VerifyResult struct {
	OK       bool     `json:"ok"`
	Trust    *float64 `json:"trust,omitempty"`
	TierUsed string   `json:"tier_used,omitempty"`
	UnitIDs  []string `json:"unit_ids,omitempty"`
	Reason   string   `json:"reason,omitempty"`
}

type IngestTraceInput struct {
	TraceID     string      `json:"trace_id,omitempty"`
	SourceAgent string      `json:"source_agent"`
	Blob        interface{} `json:"blob"`
}

type TraceResult struct {
	Status          string   `json:"status"`
	TraceID         string   `json:"trace_id"`
	PollURL         *string  `json:"poll_url,omitempty"`
	UnitIDs         []string `json:"unit_ids,omitempty"`
	StructuralCount *int     `json:"structural_count,omitempty"`
	LLMCount        *int     `json:"llm_count,omitempty"`
	SkippedCount    *int     `json:"skipped_count,omitempty"`
}

type FetchTraceResult struct {
	TraceID         string   `json:"trace_id"`
	Status          string   `json:"status"`
	BlobSize        *int64   `json:"blob_size,omitempty"`
	IngestedAt      *string  `json:"ingested_at,omitempty"`
	UnitIDs         []string `json:"unit_ids,omitempty"`
	StructuralCount *int     `json:"structural_count,omitempty"`
	LLMCount        *int     `json:"llm_count,omitempty"`
	SkippedCount    *int     `json:"skipped_count,omitempty"`
	Error           *string  `json:"error,omitempty"`
}

type IngestUnitsRequest struct {
	Units []json.RawMessage `json:"units"`
}

type SkillReviewInput struct {
	SkillName       string `json:"skill_name"`
	Diff            string `json:"diff"`
	ExternalRef     string `json:"external_ref,omitempty"`
	Context         string `json:"context,omitempty"`
	SourceAgent     string `json:"source_agent,omitempty"`
	YesLabel        string `json:"yes_label,omitempty"`
	NoLabel         string `json:"no_label,omitempty"`
	TrustedPoolOnly *bool  `json:"trusted_pool_only,omitempty"`
}

type SkillReviewResult struct {
	UnitID     string `json:"unit_id"`
	PollURL    string `json:"poll_url"`
	VerdictURL string `json:"verdict_url"`
	Created    bool   `json:"created"`
}

type ScoreResult struct {
	Counts             map[string]int `json:"counts"`
	TrustWeightedScore *float64       `json:"trust_weighted_score,omitempty"`
}

type NextUnitResult struct {
	UnitID  string                 `json:"unit_id"`
	Type    string                 `json:"type,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type SubmitJudgmentInput struct {
	UnitID  string `json:"unit_id"`
	Choice  string `json:"choice"`
	RaterID string `json:"rater_id,omitempty"`
}

type IngestUnitInput struct {
	Type        string      `json:"type"`
	Pool        string      `json:"pool,omitempty"`
	Payload     interface{} `json:"payload,omitempty"`
	ExternalRef string      `json:"external_ref,omitempty"`
	SourceAgent string      `json:"source_agent,omitempty"`
	Question    string      `json:"question,omitempty"`
}

type IngestUnitResult struct {
	OK       bool                     `json:"ok"`
	Accepted int                      `json:"accepted"`
	Rejected int                      `json:"rejected"`
	IDs      []string                 `json:"ids,omitempty"`
	Errors   []map[string]interface{} `json:"errors,omitempty"`
}

type TraceStatus struct {
	TraceID         string   `json:"trace_id"`
	Status          string   `json:"status"`
	BlobSize        *int64   `json:"blob_size,omitempty"`
	IngestedAt      *string  `json:"ingested_at,omitempty"`
	UnitIDs         []string `json:"unit_ids,omitempty"`
	StructuralCount *int     `json:"structural_count,omitempty"`
	LLMCount        *int     `json:"llm_count,omitempty"`
	SkippedCount    *int     `json:"skipped_count,omitempty"`
	Error           *string  `json:"error,omitempty"`
}

type JudgmentInput struct {
	UnitID     string      `json:"unit_id"`
	RaterID    string      `json:"rater_id"`
	Choice     string      `json:"choice"`
	LatencyMS  int         `json:"latency_ms,omitempty"`
	Confidence float64     `json:"confidence,omitempty"`
	Behavioral interface{} `json:"behavioral,omitempty"`
}

type JudgmentResult struct {
	OK                  bool    `json:"ok"`
	Token               string  `json:"token,omitempty"`
	Trust               float64 `json:"trust,omitempty"`
	TrustDelta          float64 `json:"trust_delta,omitempty"`
	EarnedCents         int     `json:"earned_cents,omitempty"`
	JudgmentsCount      int     `json:"judgments_count,omitempty"`
	DemoAgreedWithGold  bool    `json:"_demo_agreed_with_gold,omitempty"`
	DemoHoneypotFailed  bool    `json:"_demo_honeypot_failed,omitempty"`
	DemoBehavioralScore float64 `json:"_demo_behavioral_score,omitempty"`
}

type ScrubResult struct {
	Scrubbed  string                 `json:"scrubbed"`
	MappingID string                 `json:"mapping_id,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}
