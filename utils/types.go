package utils

import (
	"context"
)

type UserState struct {
	IsAuthenticated bool
	CurrentPath     string
	Username        string
	UserID          int64
	ImageUrl        string
	Email           string
	Error           string
}

type UserStateKey struct{}

func GetUserState(ctx context.Context) UserState {
	if state, ok := ctx.Value(UserStateKey{}).(UserState); ok {
		return state
	}
	return UserState{IsAuthenticated: false}
}

type UploadImageConfig struct {
	// default preview url for the image upload component
	DefaultPreviewUrl string

	//Progress is the progress precentage in 100
	Progress int

	//upload endpoint is the endpoint the form will submit on
	UploadEndpoint string

	//FormIDSelector is the id of the image upload component
	InputName string

	//Error it contains the error string if there is any error
	Error string
}
