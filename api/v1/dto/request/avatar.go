package request

// AvatarUpload contains validated avatar payload data.
type AvatarUpload struct {
	Data        []byte
	ContentType string
}

