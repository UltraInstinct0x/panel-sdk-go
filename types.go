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
	UnitID    string `json:"unit_id"`
	PollURL   string `json:"poll_url"`
	VerdictURL string `json:"verdict_url"`
	Created   bool   `json:"created"`
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
