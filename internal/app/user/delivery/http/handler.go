package http

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dgrijalva/jwt-go"
	errors "github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/errors"
	models "github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/models"
	sessionJwt "github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/session/jwt"
	customLogger "github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/tools/logger"
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/user"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
)

type UserHandler struct {
	Usecase user.Usecase
	SessionManager sessionJwt.TokenManager
}

func NewLoginHandler(storageUser user.Usecase, sessionManager sessionJwt.TokenManager) *UserHandler {
	return &UserHandler{
		Usecase: storageUser,
		SessionManager: sessionManager,
	}
}

const trace = "UserHandler"

func (handler *UserHandler) Login(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + ".Login")

	var newUserDataForInput models.UserDataForInput
	if err := ctx.Bind(&newUserDataForInput); err != nil {
		newLoginError := errors.NewError(errors.BIND_ERROR, errors.BIND_DESCR)
		logger.Error(err)
		return ctx.JSON(http.StatusBadRequest, newLoginError)
	}

	if err := ctx.Validate(&newUserDataForInput); err != nil {
		newLoginError := errors.NewError(errors.VALIDATION_ERROR, errors.VALIDATION_DESCR)
		logger.Error(err)
		return ctx.JSON(http.StatusBadRequest, newLoginError)
	}

	id, err := handler.Usecase.CheckPassword(newUserDataForInput)
	if err != nil {
		newLoginError := errors.NewError(errors.DB_ERROR, errors.BD_ERROR_DESCR)
		logger.Error(err)
		return ctx.JSON(http.StatusBadRequest, newLoginError)
	}

	if id == errors.NO_USER_ERROR {
		newLoginError := errors.NewError(errors.NO_USER_ERROR, errors.NO_USER_DESCR)
		logger.Debug(trace + " No user: " + newUserDataForInput.Name)
		return ctx.JSON(http.StatusUnauthorized, newLoginError)
	}

	if id == errors.WRONG_PASSWORD_ERROR {
		newLoginError := errors.NewError(errors.WRONG_PASSWORD_ERROR, errors.WRONG_PASSWORD_DESCR)
		logger.Debug(trace + " Wrong password for user: " + newUserDataForInput.Name)
		return ctx.JSON(http.StatusUnauthorized, newLoginError)
	}

	claimsString, err := handler.SessionManager.GetToken(id, newUserDataForInput.Name)
	if err != nil {
		newLoginError := errors.NewError(errors.TOKEN_ERROR, errors.TOKEN_ERROR_DESCR)
		logger.Error(err)
		return ctx.JSON(http.StatusInternalServerError, newLoginError)
	}

	handler.setCookieValue(ctx, claimsString)
	newUserDataForInput.Password = ""
	logger.Trace(trace + "ok login for user: " + newUserDataForInput.Name)
	return ctx.JSON(http.StatusOK, newUserDataForInput)
}

func (handler *UserHandler) SignUp(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + ".SignUp")

	var newUser models.UserDataForReg
	if err := ctx.Bind(&newUser); err != nil {
		newSignupError := errors.NewError(errors.BIND_ERROR, errors.BIND_DESCR)
		logger.Error(err)
		return ctx.JSON(http.StatusBadRequest, newSignupError)
	}

	if err := ctx.Validate(&newUser); err != nil {
		newSignupError := errors.NewError(errors.VALIDATION_ERROR, errors.VALIDATION_DESCR)
		logger.Error(err)
		return ctx.JSON(http.StatusBadRequest, newSignupError)
	}

	newId, err := handler.Usecase.AddUser(newUser)
	if err != nil {
		newSignupError := errors.NewError(newId, err.Error())
		logger.Error(err)
		return ctx.JSON(http.StatusInternalServerError, newSignupError)
	}

	if newId == errors.USER_EXISTS_ERROR {
		newSignupError := errors.NewError(errors.USER_EXISTS_ERROR, errors.USER_EXISTS_DESCR)
		logger.Debug(trace + " User Already exist: " + newUser.Name)
		return ctx.JSON(http.StatusUnauthorized, newSignupError)
	}

	claimsString, err := handler.SessionManager.GetToken(newId, newUser.Name)
	if err != nil {
		newSignupError := errors.NewError(errors.TOKEN_ERROR, errors.TOKEN_ERROR_DESCR)
		logger.Error(err)
		return ctx.JSON(http.StatusInternalServerError, newSignupError)
	}
	handler.setCookieValue(ctx, claimsString)

	newUser.Password = ""
	logger.Trace(trace + "ok signup for user: " + newUser.Name)
	return ctx.JSON(http.StatusOK, newUser)
}

func (handler *UserHandler) UploadAvatar(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + ".UploadAvatar")

	idNum, err := handler.SessionManager.ParseTokenFromContext(ctx)
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusUnauthorized, errors.NewError(errors.TOKEN_ERROR, errors.TOKEN_ERROR_DESCR))
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	src, err := file.Open()
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	defer src.Close()

	size := file.Size
	buffer := make([]byte, size)

	_, err = src.Read(buffer)
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	fileBytes := bytes.NewReader(buffer)

	fileName := handler.Usecase.GenerateAvatarName()

	bucket := "products-bucket-ozon-good-vibes"

	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(
		&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fileName),
			Body:   fileBytes,
		})

	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	err = handler.Usecase.SaveAvatarName(int(idNum), fileName)
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	logger.Trace("success upload avatar")
	return ctx.HTML(http.StatusOK, fileName)
}

func (handler *UserHandler) Profile(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + ".Profile")

	token := ctx.Get("token").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	idString := claims["id"].(string)
	idNum, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusUnauthorized, errors.NewError(errors.TOKEN_ERROR, errors.TOKEN_ERROR_DESCR))
	}

	userData, err := handler.Usecase.GetUserDataByID(idNum)
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusUnauthorized, errors.NewError(errors.TOKEN_ERROR, err.Error()))
	}

	return ctx.JSON(http.StatusOK, userData)
}

func (handler *UserHandler) Logout(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + ".Logout")

	cookie := &http.Cookie{
		Name:     "session_id",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteNoneMode,
		//Secure:   true,
	}
	ctx.SetCookie(cookie)
	return ctx.NoContent(http.StatusOK)
}

func (handler *UserHandler) setCookieValue(ctx echo.Context, value string) {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + ".setCookieValue")

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    value,
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 72),
		SameSite: http.SameSiteNoneMode,
		//Secure:   true,
	}

	ctx.SetCookie(cookie)
}
