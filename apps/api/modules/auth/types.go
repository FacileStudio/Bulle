package auth

type AssistantProfileResponse struct {
	UserID    int64    `json:"user_id,omitempty"`
	Name      string   `json:"name"`
	AvatarURL string   `json:"avatar_url"`
	Archetype string   `json:"archetype"`
	Traits    []string `json:"traits"`
}

type UserResponse struct {
	ID        int64                     `json:"id"`
	Email     string                    `json:"email"`
	Name      string                    `json:"name"`
	AvatarURL string                    `json:"avatar_url"`
	Locale    string                    `json:"locale"`
	Plan      string                    `json:"plan,omitempty"`
	CreatedAt string                    `json:"created_at"`
	Assistant *AssistantProfileResponse `json:"assistant,omitempty"`
}

type MeResponse struct {
	User UserResponse `json:"user"`
}
