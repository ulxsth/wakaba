package handler

// WorkerRequest is a unified payload for the async worker lambda.
type WorkerRequest struct {
	// Common fields
	Type             string `json:"type"` // "command" or "component"
	InteractionID    string `json:"interaction_id"`
	InteractionToken string `json:"interaction_token"`
	ChannelID        string `json:"channel_id"`
	ApplicationID    string `json:"application_id"`
	GuildID          string `json:"guild_id"`

	// For Commands
	CommandName string          `json:"command_name,omitempty"`
	CommandArgs map[string]any  `json:"command_args,omitempty"`

	// For Components (Buttons)
	CustomID string `json:"custom_id,omitempty"`
}

// Command Arguments structures
type SummarizeArgs struct {
	DateArg   string `json:"date_arg"`
	WithTitle bool   `json:"with_title"`
}

type TodoListArgs struct {
	SubCommand string `json:"sub_command"`
	Content    string `json:"content"`
}
