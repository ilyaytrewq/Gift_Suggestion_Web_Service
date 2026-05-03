package http

type recommendRequest struct {
	Occasion             string   `json:"occasion"`
	Relationship         string   `json:"relationship"`
	RecipientAge         *int     `json:"recipient_age,omitempty"`
	RecipientGender      *string  `json:"recipient_gender,omitempty"`
	BudgetMax            string   `json:"budget_max"`
	PreferredCategoryIDs []string `json:"preferred_category_ids,omitempty"`
	Interests            []string `json:"interests,omitempty"`
	TopN                 int      `json:"top_n,omitempty"`
	UseWishlistContext   *bool    `json:"use_wishlist_context,omitempty"`
}
