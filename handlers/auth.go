package handlers

import (
	"http-proxy/conf"
	"http-proxy/logger"
	"http-proxy/models"
	"http-proxy/utils"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()

type AuthHandler struct{}

func (ah AuthHandler) Authenticate(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	var body models.AuthSchema
	if err := c.BodyParser(&body); err != nil {
		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}

	var conn *memphis.Conn
	var err error
	if configuration.CLIENT_CERT_PATH != "" && configuration.CLIENT_KEY_PATH != "" && configuration.ROOT_CA_PATH != "" {
		conn, err = memphis.Connect(
			configuration.MEMPHIS_HOST,
			body.Username,
			body.ConnectionToken,
			memphis.Tls(configuration.CLIENT_CERT_PATH, configuration.CLIENT_KEY_PATH, configuration.ROOT_CA_PATH),
		)
	} else {
		conn, err = memphis.Connect(configuration.MEMPHIS_HOST, body.Username, body.ConnectionToken)
	}

	if err != nil {
		if strings.Contains(err.Error(), "Authorization Violation") {
			log.Warnf("Authentication error")
			return c.Status(401).JSON(fiber.Map{
				"message": "Wrong credentials",
			})
		}

		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}
	conn.Close()
	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins)
	if err != nil {
		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Create tokens error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}

func createTokens(tokenExpiryMins, refreshTokenExpiryMins int) (string, string, int, int, error) {
	if tokenExpiryMins <= 0 {
		tokenExpiryMins = configuration.JWT_EXPIRES_IN_MINUTES
	}

	if refreshTokenExpiryMins <= 0 {
		refreshTokenExpiryMins = configuration.JWT_EXPIRES_IN_MINUTES
	}

	atClaims := jwt.MapClaims{}
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenExpiryMins)).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(configuration.JWT_SECRET))
	if err != nil {
		return "", "", 0, 0, err
	}

	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(refreshTokenExpiryMins)).Unix()
	at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	refreshToken, err := at.SignedString([]byte(configuration.REFRESH_JWT_SECRET))
	if err != nil {
		return "", "", 0, 0, err
	}
	return token, refreshToken, tokenExpiryMins, refreshTokenExpiryMins, nil
}

func (ah AuthHandler) RefreshToken(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	var body models.RefreshTokenSchema
	if err := c.BodyParser(&body); err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}

	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins)
	if err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Create tokens error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}
