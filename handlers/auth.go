package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"pong-htmx/config"
	"pong-htmx/models"
	"pong-htmx/repository"
	"pong-htmx/utils"
	"pong-htmx/views/components"
	"pong-htmx/views/pages"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	env           *config.Env
	googleConfig  *oauth2.Config
	userRepo      *repository.UserRepository
	sessionRepo   *repository.SessionsRepository
	magicLinkRepo *repository.MagicLinkTokenRepository
	awsS3Client   *s3.Client
	awsSesClient  *ses.Client
	mLog          *config.ManiaLogger
}

func NewAuthHandler(googleConfig *oauth2.Config, userRepo *repository.UserRepository, sessionRepo *repository.SessionsRepository, magicLinkRepo *repository.MagicLinkTokenRepository, awsS3Client *s3.Client, awsSesClient *ses.Client, env *config.Env, mLog *config.ManiaLogger) *AuthHandler {
	return &AuthHandler{
		env:           env,
		googleConfig:  googleConfig,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		magicLinkRepo: magicLinkRepo,
		awsS3Client:   awsS3Client,
		awsSesClient:  awsSesClient,
		mLog:          mLog,
	}
}

func (h *AuthHandler) GoogleLogin(ctx echo.Context) error {
	url := h.googleConfig.AuthCodeURL("randomState", oauth2.AccessTypeOffline)
	return ctx.Redirect(http.StatusSeeOther, url)
}

func (h *AuthHandler) GoogleLoginCallback(ctx echo.Context) error {
	code := ctx.QueryParam("code")
	if code == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to get query 'code'"})
	}

	token, err := h.googleConfig.Exchange(ctx.Request().Context(), code)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to exchange token"})
	}

	client := h.googleConfig.Client(ctx.Request().Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to get user info"})
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode user info"})
	}

	current_time := time.Now()
	dbUser := models.User{
		Email:      userInfo.Email,
		Username:   userInfo.Email,
		Provider:   utils.GOOGLE_PROVIDER,
		ImageUrl:   userInfo.Picture,
		IsActive:   true,
		IsVerified: true,
		CreatedAt:  &current_time,
		UpdatedAt:  nil,
	}

	user, err := h.userRepo.GetByEmail(userInfo.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			user, err = h.userRepo.Create(dbUser)
			if err != nil {
				log.Println(err)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Cannot create user"})
			}
		} else {
			log.Println(err)
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Couldn't find user"})
		}
	}

	if user.Provider == utils.EMAIL_PROVIDER {
		return utils.Render(ctx, http.StatusBadRequest, pages.Login(ctx.Request().Context(), false, "", utils.TemplateError{ErrorMessage: "User already exists with email provider. Please login with email"}))
	}

	sessionToken, err := utils.GenerateSessionToken()
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate session token"})
	}

	// Create a new session in the database
	expiresAt := time.Now().Add(24 * time.Hour) // Set sessionRecord expiration to 24 hours from now
	sessionRecord := models.Session{
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: &expiresAt,
	}

	err = h.sessionRepo.Create(sessionRecord)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
	}

	// Set session cookie
	cookie := new(http.Cookie)
	cookie.Name = "session_token"
	cookie.Value = sessionToken
	cookie.Expires = expiresAt
	cookie.Path = "/"
	cookie.HttpOnly = true
	//cookie.Secure = true
	cookie.SameSite = http.SameSiteLaxMode
	ctx.SetCookie(cookie)

	return ctx.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("session_token")
	if err == nil {
		// Remove from database
		err = h.sessionRepo.Delete(cookie.Value)
		if err != nil {
			// Log the error, but continue
			log.Printf("Error deleting session: %v", err)
		}

		// Clear the cookie
		newCookie := &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			//Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(newCookie)
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}

func (h *AuthHandler) SendMagicLink(c echo.Context) error {
	email := c.FormValue("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}

	token, err := utils.GenerateSessionToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}
	// Save the token in the magiclink repository with an expiration time of 3 minutes
	expiresAt := time.Now().Add(20 * time.Minute)
	err = h.magicLinkRepo.CreateMagicLinkToken(email, token, expiresAt)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create magic link token"})
	}

	err = config.SendMagicLink(email, h.env.SES_EMAIL, token, h.awsSesClient)
	if err != nil {
		fmt.Println(err)
		h.mLog.LogError("MAGIC_LINK_EMAIL_FAIL", err, &c)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email"})
	}

	returnHTML := `<p class="text-md text-green-600 mx-auto">Login link sent. Check your email.</p>`
	return c.HTML(http.StatusOK, returnHTML)
}

func (h *AuthHandler) VerifyMagicLinkTokenPage(ctx echo.Context) error {
	token := ctx.QueryParam("token")
	if token == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Token is required"})
	}
	loginUrl := fmt.Sprintf("/auth/magic/login/%s", token)
	return utils.Render(ctx, http.StatusOK, pages.Login(ctx.Request().Context(), true, loginUrl, utils.TemplateError{}))
}

func (h *AuthHandler) LoginWithMagicLink(ctx echo.Context) error {
	token := ctx.Param("token")
	if token == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Token is required"})
	}

	magicLinkRecord, err := h.magicLinkRepo.GetByToken(token)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid token"})
	}

	if magicLinkRecord.ExpiresAt.Before(time.Now()) {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Token has expired"})
	}

	user, err := h.userRepo.GetByEmail(magicLinkRecord.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			createdAt := time.Now()
			user = models.User{
				Email:      magicLinkRecord.Email,
				Username:   magicLinkRecord.Email,
				Provider:   utils.EMAIL_PROVIDER,
				IsActive:   true,
				IsVerified: true,
				CreatedAt:  &createdAt,
			}
			user, err = h.userRepo.Create(user)
			if err != nil {
				fmt.Println(err)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Cannot create user"})
			}
		} else {
			log.Println(err)
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Couldn't find user"})
		}
	}

	sessionToken, err := utils.GenerateSessionToken()
	if err != nil {
		fmt.Println(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate session token"})
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	sessionRecord := models.Session{
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: &expiresAt,
	}

	err = h.sessionRepo.Create(sessionRecord)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
	}

	// Set session cookie
	cookie := new(http.Cookie)
	cookie.Name = "session_token"
	cookie.Value = sessionToken
	cookie.Expires = expiresAt
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	ctx.SetCookie(cookie)

	return ctx.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) UpdateProfile(ctx echo.Context) error {
	username := ctx.FormValue("username")
	user := ctx.Get("user").(models.User)
	editUser, err := h.userRepo.GetByUsername(user.Username)
	if err != nil {
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileContent(editUser))
	}
	editUser.Username = username
	fmt.Printf("%+v\n", editUser)
	err = h.userRepo.Update(editUser)
	if err != nil {
		fmt.Printf("%v", err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileContent(editUser))

	}
	updatedUser, err := h.userRepo.GetByID(user.ID)
	if err != nil {
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileContent(editUser))
	}
	return utils.Render(ctx, http.StatusOK, components.ProfileContent(updatedUser))
}

func (h *AuthHandler) UpdateProfileImage(ctx echo.Context) error {
	profileImage, err := ctx.FormFile("profile_image")
	if err != nil {
		fmt.Println(err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileEditImage(utils.UploadImageConfig{
			DefaultPreviewUrl: "/static/assets/AVATAR.svg",
			Progress:          0,
			UploadEndpoint:    "/user/profile/image",
			InputName:         "profile_image",
			Error:             err.Error(),
		}, utils.GetUserState(ctx.Request().Context())))
	}
	user := ctx.Get("user").(models.User)

	foundUser, err := h.userRepo.GetByID(user.ID)
	if err != nil {
		fmt.Println(err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileEditImage(utils.UploadImageConfig{
			DefaultPreviewUrl: "/static/assets/AVATAR.svg",
			Progress:          0,
			UploadEndpoint:    "/user/profile/image",
			InputName:         "profile_image",
			Error:             err.Error(),
		}, utils.GetUserState(ctx.Request().Context())))
	}

	h.awsS3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(h.env.PROFILE_IMAGE_S3_BUCKET),
		Key:    &foundUser.ImageUrl,
	})

	fmt.Printf("%+v", profileImage) // fine till here
	imageExtension := utils.FileExtension(profileImage.Filename)
	if imageExtension == "" {
		fmt.Println(err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileEditImage(utils.UploadImageConfig{
			DefaultPreviewUrl: "/static/assets/AVATAR.svg",
			Progress:          0,
			UploadEndpoint:    "/user/profile/image",
			InputName:         "profile_image",
			Error:             "invalid file type",
		}, utils.GetUserState(ctx.Request().Context())))
	}
	newFileName := utils.NewProfileFileName(imageExtension)
	trx, ok := ctx.Get("db_trx").(*sql.Tx)
	if !ok {
		fmt.Println(err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileEditImage(utils.UploadImageConfig{
			DefaultPreviewUrl: "/static/assets/AVATAR.svg",
			Progress:          0,
			UploadEndpoint:    "/user/profile/image",
			InputName:         "profile_image",
			Error:             "database transaction error",
		}, utils.GetUserState(ctx.Request().Context())))

	}
	err = h.userRepo.UpdateProfileImage(user.ID, newFileName, trx)
	if err != nil {
		fmt.Println(err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileEditImage(utils.UploadImageConfig{
			DefaultPreviewUrl: "/static/assets/AVATAR.svg",
			Progress:          0,
			UploadEndpoint:    "/user/profile/image",
			InputName:         "profile_image",
			Error:             err.Error(),
		}, utils.GetUserState(ctx.Request().Context())))
	}
	imageFile, err := profileImage.Open()
	if err != nil {
		fmt.Println(err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileEditImage(utils.UploadImageConfig{
			DefaultPreviewUrl: "/static/assets/AVATAR.svg",
			Progress:          0,
			UploadEndpoint:    "/user/profile/image",
			InputName:         "profile_image",
			Error:             err.Error(),
		}, utils.GetUserState(ctx.Request().Context())))

	}
	defer imageFile.Close()

	imageBody, err := io.ReadAll(imageFile)
	if err != nil {
		fmt.Println(err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileEditImage(utils.UploadImageConfig{
			DefaultPreviewUrl: "/static/assets/AVATAR.svg",
			Progress:          0,
			UploadEndpoint:    "/user/profile/image",
			InputName:         "profile_image",
			Error:             err.Error(),
		}, utils.GetUserState(ctx.Request().Context())))
	}
	fileSize := profileImage.Size
	fileType := http.DetectContentType(imageBody)
	fmt.Println(fileType)
	_, err = h.awsS3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(h.env.PROFILE_IMAGE_S3_BUCKET),
		Key:           &newFileName,
		Body:          bytes.NewReader(imageBody),
		ContentLength: aws.Int64(fileSize),
		ContentType:   aws.String(fileType),
		Tagging:       aws.String("ProfileImage="),
	})
	if err != nil {
		fmt.Println(err)
		return utils.Render(ctx, http.StatusInternalServerError, components.ProfileEditImage(utils.UploadImageConfig{
			DefaultPreviewUrl: "/static/assets/AVATAR.svg",
			Progress:          0,
			UploadEndpoint:    "/user/profile/image",
			InputName:         "profile_image",
			Error:             err.Error(),
		}, utils.GetUserState(ctx.Request().Context())))
	}
	return nil
}

func (h *AuthHandler) ProfileImage(ctx echo.Context) error {
	imageKey := ctx.QueryParam("image_key")
	if imageKey == "" {
		return ctx.String(http.StatusInternalServerError, "no image key provided")
	}
	user := ctx.Get("user").(models.User)
	if user.ImageUrl != imageKey {
		return ctx.String(http.StatusInternalServerError, "invalid image key")
	}
	out, err := h.awsS3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(h.env.PROFILE_IMAGE_S3_BUCKET),
		Key:    &imageKey,
	})
	if err != nil {
		return ctx.String(http.StatusNotFound, "image not found")
	}
	defer out.Body.Close()
	imageBody, err := io.ReadAll(out.Body)
	if err != nil {
		fmt.Println(err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	fmt.Println(imageBody)
	fileType := out.Metadata["Content-Type"]
	return ctx.Blob(http.StatusOK, fileType, imageBody)
}
