package http

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/errors"
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/models"
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/product"
	sessionJwt "github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/session/jwt"
	customLogger "github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/tools/logger"
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/tools/sanitizer"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

const BucketUrl = ""

type ProductHandler struct {
	useCase        product.UseCase
	SessionManager sessionJwt.TokenManager
}

func NewProductHandler(useCase product.UseCase, sessionManager sessionJwt.TokenManager) *ProductHandler {
	return &ProductHandler{
		useCase:        useCase,
		SessionManager: sessionManager,
	}
}

const trace = "ProductHandler"

//это должно быть, пока не нужно
/*
func addProduct(ctx echo.Context) error {
	return nil
}*/

//пока решаем вопросы с пагинацией - так

func (ph *ProductHandler) AddProduct(ctx echo.Context) error {
	var newProduct models.Product
	if err := ctx.Bind(&newProduct); err != nil {
		newError := errors.NewError(errors.BIND_ERROR, errors.BIND_DESCR)
		return ctx.JSON(http.StatusBadRequest, newError)
	}

	if err := ctx.Validate(&newProduct); err != nil {
		newError := errors.NewError(errors.VALIDATION_ERROR, errors.VALIDATION_DESCR)
		return ctx.JSON(http.StatusBadRequest, newError)
	}

	newProduct = sanitizer.SanitizeData(&newProduct).(models.Product)

	productId, err := ph.useCase.AddProduct(newProduct)
	if err != nil {
		newError := errors.NewError(errors.SERVER_ERROR, errors.BD_ERROR_DESCR)
		return ctx.JSON(http.StatusInternalServerError, newError)
	}

	newProduct.Id = productId

	return ctx.JSON(http.StatusOK, newProduct)
}

func (ph *ProductHandler) AddFavouriteProduct(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + "AddFavouriteProduct")

	idNum, err := ph.SessionManager.ParseTokenFromContext(ctx.Request().Context())
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusUnauthorized, errors.NewError(errors.TOKEN_ERROR, errors.TOKEN_ERROR_DESCR))
	}

	var newProduct models.FavouriteProduct
	if err := ctx.Bind(&newProduct); err != nil {
		newError := errors.NewError(errors.BIND_ERROR, errors.BIND_DESCR)
		return ctx.JSON(http.StatusBadRequest, newError)
	}

	if err := ctx.Validate(&newProduct); err != nil {
		newError := errors.NewError(errors.VALIDATION_ERROR, errors.VALIDATION_DESCR)
		return ctx.JSON(http.StatusBadRequest, newError)
	}

	newProduct = sanitizer.SanitizeData(&newProduct).(models.FavouriteProduct)
	newProduct.UserId = int(idNum)

	err = ph.useCase.AddFavouriteProduct(newProduct)
	if err != nil {
		newError := errors.NewError(errors.SERVER_ERROR, errors.BD_ERROR_DESCR)
		return ctx.JSON(http.StatusInternalServerError, newError)
	}

	return ctx.JSON(http.StatusOK, newProduct)
}

func (ph *ProductHandler) DeleteFavouriteProduct(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + ".DeleteFavouriteProduct")

	userId, err := ph.SessionManager.ParseTokenFromContext(ctx.Request().Context())
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusUnauthorized, errors.NewError(errors.TOKEN_ERROR, errors.TOKEN_ERROR_DESCR))
	}

	var deleteProduct models.FavouriteProduct

	if err := ctx.Bind(&deleteProduct); err != nil {
		logger.Error(err)
		newProductError := errors.NewError(errors.BIND_ERROR, errors.BIND_DESCR)
		return ctx.JSON(http.StatusBadRequest, newProductError)
	}

	if err := ctx.Validate(&deleteProduct); err != nil {
		logger.Error(err, deleteProduct)
		newProductError := errors.NewError(errors.VALIDATION_ERROR, errors.VALIDATION_DESCR)
		return ctx.JSON(http.StatusBadRequest, newProductError)
	}

	deleteProduct.UserId = int(userId)

	err = ph.useCase.DeleteFavouriteProduct(deleteProduct)
	if err != nil {
		logger.Error(err, deleteProduct)
		newProductError := errors.NewError(errors.SERVER_ERROR, err.Error())
		return ctx.JSON(http.StatusInternalServerError, newProductError)
	}

	logger.Trace(trace + " success DeleteFavouriteProduct")
	return ctx.JSON(http.StatusOK, deleteProduct)
}

func (ph *ProductHandler) GetAllProducts(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + "GetAllProducts")

	answer, err := ph.useCase.GetAllProducts()
	if err != nil {
		logger.Error(err)
		newProductError := errors.NewError(errors.DB_ERROR, errors.BD_ERROR_DESCR)
		return ctx.JSON(http.StatusInternalServerError, newProductError)
	}

	for i, _ := range answer {
		answer[i] = sanitizer.SanitizeData(&answer[i]).(models.Product)
	}

	return ctx.JSON(http.StatusOK, answer)
}

func (ph *ProductHandler) GetFavouriteProducts(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + "GetFavouriteProducts")

	idNum, err := ph.SessionManager.ParseTokenFromContext(ctx.Request().Context())
	if err != nil {
		logger.Error(err)
		return ctx.JSON(http.StatusUnauthorized, errors.NewError(errors.TOKEN_ERROR, errors.TOKEN_ERROR_DESCR))
	}

	products, err := ph.useCase.GetFavouriteProducts(int(idNum))
	if err != nil {
		logger.Error(err)
		newProductError := errors.NewError(errors.DB_ERROR, errors.BD_ERROR_DESCR)
		return ctx.JSON(http.StatusInternalServerError, newProductError)
	}

	for i, _ := range products {
		products[i] = sanitizer.SanitizeData(&products[i]).(models.Product)
	}

	return ctx.JSON(http.StatusOK, products)
}

func (ph *ProductHandler) GetProductById(ctx echo.Context) error {
	logger := customLogger.TryGetLoggerFromContext(ctx)
	logger.Trace(trace + "GetProductById")

	val := ctx.QueryParams()
	idString := val.Get("id")
	if idString == "" {
		logger.Error("bad query param for GetProductById")
		newProductError := errors.NewError(errors.VALIDATION_ERROR, errors.VALIDATION_DESCR)
		return ctx.JSON(http.StatusBadRequest, newProductError)
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		logger.Error(err)
		newProductError := errors.NewError(errors.VALIDATION_ERROR, errors.VALIDATION_DESCR)
		return ctx.JSON(http.StatusBadRequest, newProductError)
	}

	answer, err := ph.useCase.GetProductById(id)
	if err != nil {
		logger.Error(err)
		newProductError := errors.NewError(errors.DB_ERROR, err.Error())
		return ctx.JSON(http.StatusBadRequest, newProductError)
	}

	answer = sanitizer.SanitizeData(&answer).(models.Product)

	return ctx.JSON(http.StatusOK, answer)
}

func (ph *ProductHandler) UploadProduct(ctx echo.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	src, err := file.Open()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	defer src.Close()

	size := file.Size
	buffer := make([]byte, size)

	_, err = src.Read(buffer)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	fileBytes := bytes.NewReader(buffer)

	fileName := ph.useCase.GenerateProductImageName()

	bucket := ""

	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(
		&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fileName),
			Body:   fileBytes,
		})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	productId, err := strconv.Atoi(ctx.FormValue("product_id"))

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	err = ph.useCase.SaveProductImageName(productId, BucketUrl+fileName)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	return ctx.HTML(http.StatusOK, BucketUrl+fileName)
}
