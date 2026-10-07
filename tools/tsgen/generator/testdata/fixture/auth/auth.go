package auth

import msg "fixture/api"

const KeyEmailTaken msg.Key = "auth.email_taken"

const Quoted msg.Key = `it's`

const cookie = "session"

//tsgen:assets/app/Auth/types/SignIn.ts SignIn request
type SignIn struct {
	Email string `json:"email" binding:"required,email"`

	Password string `json:"password" binding:"required"`

	Remember bool `json:"default"`
}
