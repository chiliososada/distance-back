package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chiliososada/distance-back/config"
	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/pkg/logger"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/storage"
	"google.golang.org/api/option"
)

var (
	firebaseAuth    *auth.Client
	firebaseStorage *storage.Client
	firebaseApp     *firebase.App
)

// GetStorage 获取 Storage 客户端
func GetStorageClient() *storage.Client {
	return firebaseStorage
}

// InitFirebase 初始化Firebase所有服务
func InitFirebase(cfg *config.FirebaseConfig) error {
	ctx := context.Background()
	opt := option.WithCredentialsFile(cfg.CredentialsFile)

	app, err := firebase.NewApp(ctx, &firebase.Config{
		StorageBucket: cfg.StorageBucket,
	}, opt)
	if err != nil {
		return fmt.Errorf("error initializing firebase app: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("error initializing firebase auth: %v", err)
	}

	storageClient, err := app.Storage(ctx)
	if err != nil {
		return fmt.Errorf("error initializing firebase storage: %v", err)
	}

	firebaseApp = app
	firebaseAuth = authClient
	firebaseStorage = storageClient

	logger.Info("Firebase initialized successfully")
	return nil
}

// GetUserByUID 通过Firebase UID获取用户信息
func GetUserByUID(ctx context.Context, uid string) (*auth.UserRecord, error) {
	if firebaseAuth == nil {
		return nil, fmt.Errorf("firebase auth client not initialized")
	}

	user, err := firebaseAuth.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error getting user by UID: %v", err)
	}

	return user, nil
}

// GetUserByEmail 通过邮箱获取用户信息
func GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error) {
	if firebaseAuth == nil {
		return nil, fmt.Errorf("firebase auth client not initialized")
	}

	user, err := firebaseAuth.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user by email: %v", err)
	}

	return user, nil
}

// GetUserByPhone 通过手机号获取用户信息
func GetUserByPhone(ctx context.Context, phone string) (*auth.UserRecord, error) {
	if firebaseAuth == nil {
		return nil, fmt.Errorf("firebase auth client not initialized")
	}

	user, err := firebaseAuth.GetUserByPhoneNumber(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("error getting user by phone: %v", err)
	}

	return user, nil
}

// CreateCustomToken 创建自定义令牌
func CreateCustomToken(ctx context.Context, uid string, claims map[string]interface{}) (string, error) {
	if firebaseAuth == nil {
		return "", fmt.Errorf("firebase auth client not initialized")
	}

	// 如果有 claims，使用 CustomTokenWithClaims
	if claims != nil {
		token, err := firebaseAuth.CustomTokenWithClaims(ctx, uid, claims)
		if err != nil {
			return "", fmt.Errorf("error creating custom token with claims: %v", err)
		}
		return token, nil
	}

	// 如果没有 claims，使用 CustomToken
	token, err := firebaseAuth.CustomToken(ctx, uid)
	if err != nil {
		return "", fmt.Errorf("error creating custom token: %v", err)
	}

	return token, nil
}

// UpdateUserClaims 更新用户自定义声明
func UpdateUserClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	if firebaseAuth == nil {
		return fmt.Errorf("firebase auth client not initialized")
	}

	if err := firebaseAuth.SetCustomUserClaims(ctx, uid, claims); err != nil {
		return fmt.Errorf("error updating user claims: %v", err)
	}

	return nil
}

// DisableUser 禁用用户
func DisableUser(ctx context.Context, uid string) error {
	if firebaseAuth == nil {
		return fmt.Errorf("firebase auth client not initialized")
	}

	update := (&auth.UserToUpdate{}).
		Disabled(true)

	if _, err := firebaseAuth.UpdateUser(ctx, uid, update); err != nil {
		return fmt.Errorf("error disabling user: %v", err)
	}

	return nil
}

// EnableUser 启用用户
func EnableUser(ctx context.Context, uid string) error {
	if firebaseAuth == nil {
		return fmt.Errorf("firebase auth client not initialized")
	}

	update := (&auth.UserToUpdate{}).
		Disabled(false)

	if _, err := firebaseAuth.UpdateUser(ctx, uid, update); err != nil {
		return fmt.Errorf("error enabling user: %v", err)
	}

	return nil
}

// VerifyIDToken 验证Firebase ID令牌
func VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	if firebaseAuth == nil {
		return nil, fmt.Errorf("firebase auth client not initialized")
	}

	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, fmt.Errorf("empty id token")
	}

	token, err := firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying ID token: %v", err)
	}

	return token, nil
}

func SessionCookie(ctx context.Context, idToken string, expiresIn time.Duration) (string, error) {
	return firebaseAuth.SessionCookie(ctx, idToken, expiresIn)
}

func VeirfySessionCookie(ctx context.Context, session string) (*auth.Token, error) {
	return firebaseAuth.VerifySessionCookie(ctx, session)
}

func RevokeSession(ctx context.Context, uid string) error {
	return firebaseAuth.RevokeRefreshTokens(ctx, uid)
}

func UpdateUserProfile(ctx context.Context, current_session *SessionData, request *request.UpdateProfileRequest) (SessionData, *auth.UserRecord, error) {
	var utu auth.UserToUpdate
	if request.Nickname != "" {
		utu = *utu.DisplayName(request.Nickname)
	}

	if request.AvatarURL != "" {
		utu = *utu.PhotoURL(request.AvatarURL)
	}

	if request.Nickname == "" && request.AvatarURL == "" {
		return *current_session, nil, nil
	}

	rec, err := firebaseAuth.UpdateUser(ctx, current_session.UID, &utu)
	updatedSession := SessionData{
		LoginInfo: response.LoginInfo{
			CsrfToken:   current_session.CsrfToken,
			UID:         rec.UID,
			DisplayName: rec.DisplayName,
			PhotoUrl:    rec.PhotoURL,
			Email:       rec.Email,
		},
		Cookie: current_session.Cookie,
	}
	return updatedSession, rec, err

}
