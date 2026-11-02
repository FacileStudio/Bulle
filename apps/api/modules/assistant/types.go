package assistant

type ProfileResponse struct {
	UserID    int64    `json:"user_id"`
	Name      string   `json:"name"`
	AvatarURL string   `json:"avatar_url"`
	Archetype string   `json:"archetype"`
	Traits    []string `json:"traits"`
}

type UpdateProfileRequest struct {
	Name      string   `json:"name"`
	AvatarURL string   `json:"avatar_url"`
	Archetype string   `json:"archetype"`
	Traits    []string `json:"traits"`
}
