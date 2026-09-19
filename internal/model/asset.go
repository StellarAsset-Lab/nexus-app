package model

// Asset is the projection of a Registry AssetRecord.
type Asset struct {
	Asset             string `json:"asset"`
	Issuer            string `json:"issuer"`
	Active            bool   `json:"active"`
	FirstSeenLedger   int64  `json:"firstSeenLedger"`
	LastUpdatedLedger int64  `json:"lastUpdatedLedger"`
}

// Distribution is the projection of a Registry DistributionConfig.
type Distribution struct {
	Asset                string `json:"asset"`
	Distributor          string `json:"distributor"`
	EligibilityAuthority string `json:"eligibilityAuthority"`
	Active               bool   `json:"active"`
	FirstSeenLedger      int64  `json:"firstSeenLedger"`
	LastUpdatedLedger    int64  `json:"lastUpdatedLedger"`
}

// Eligibility is the projection of a Registry EligibilityRecord. It is a
// projection only — the authoritative eligibility check is the Registry
// contract's own is_eligible read, not this table.
type Eligibility struct {
	Asset             string `json:"asset"`
	Distributor       string `json:"distributor"`
	Buyer             string `json:"buyer"`
	ValidUntilLedger  int64  `json:"validUntilLedger"`
	Active            bool   `json:"active"`
	LastUpdatedLedger int64  `json:"lastUpdatedLedger"`
}
