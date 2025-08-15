package ads

type Ad struct {
	AdID           string   `json:"adId"`
	Title          string   `json:"title"`
	GameFamily     string   `json:"gameFamily"`
	TargetAudience []string `json:"targetAudience"`
	Priority       int      `json:"priority"`
	CreatedAt      string   `json:"createdAt"`
	MaxWaitTime    int      `json:"maxWaitTime"`
}
