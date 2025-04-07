package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/jmoloko/taskmange/internal/domain/models"
)

func TestAuth(t *testing.T) {
	env, cleanup := SetupTestEnv(t)
	defer cleanup()

	t.Run("Registration", func(t *testing.T) {
		cases := []struct {
			name     string
			payload  RegisterRequest
			wantCode int
		}{
			{
				name: "valid_registration",
				payload: RegisterRequest{
					Email:    "test@example.com",
					Password: "password123",
				},
				wantCode: http.StatusCreated,
			},
			{
				name: "invalid_email",
				payload: RegisterRequest{
					Email:    "invalid-email",
					Password: "password123",
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name: "empty_password",
				payload: RegisterRequest{
					Email:    "test@example.com",
					Password: "",
				},
				wantCode: http.StatusBadRequest,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				// Очищаем данные перед каждым тестом
				require.NoError(t, clearTestData(env))

				resp, err := makeRequest(env, "POST", "/api/auth/register", tc.payload, "")
				require.NoError(t, err)
				assert.Equal(t, tc.wantCode, resp.StatusCode)

				if tc.wantCode == http.StatusOK {
					// Проверяем что пользователь создался в БД
					var user struct{ ID string }
					err := env.DB.QueryRow("SELECT id FROM users WHERE email = $1", tc.payload.Email).Scan(&user.ID)
					require.NoError(t, err)
					assert.NotEmpty(t, user.ID)
				}
			})
		}
	})

	t.Run("Login", func(t *testing.T) {
		// Создаем тестового пользователя
		validUser := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		resp, err := makeRequest(env, "POST", "/api/auth/register", validUser, "")
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		cases := []struct {
			name     string
			payload  LoginRequest
			wantCode int
		}{
			{
				name: "valid_credentials",
				payload: LoginRequest{
					Email:    validUser.Email,
					Password: validUser.Password,
				},
				wantCode: http.StatusOK,
			},
			{
				name: "invalid_password",
				payload: LoginRequest{
					Email:    validUser.Email,
					Password: "wrong_password",
				},
				wantCode: http.StatusUnauthorized,
			},
			{
				name: "non_existent_user",
				payload: LoginRequest{
					Email:    "nonexistent@example.com",
					Password: "password123",
				},
				wantCode: http.StatusUnauthorized,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := makeRequest(env, "POST", "/api/auth/login", tc.payload, "")
				require.NoError(t, err)
				assert.Equal(t, tc.wantCode, resp.StatusCode)

				if tc.wantCode == http.StatusOK {
					var loginResp LoginResponse
					err := json.NewDecoder(resp.Body).Decode(&loginResp)
					require.NoError(t, err)
					assert.NotEmpty(t, loginResp.Token)
				}
			})
		}
	})
}

func TestAuth_Register(t *testing.T) {
	env, cleanup := SetupTestEnv(t)
	defer cleanup()

	// Очищаем базу данных перед тестом
	err := clearTestData(env)
	require.NoError(t, err)

	// Регистрация нового пользователя
	req := models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := makeRequest(env, "POST", "/api/auth/register", req, "")
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Проверяем, что пароль захеширован с помощью bcrypt
	var passwordHash string
	err = env.DB.QueryRow("SELECT password_hash FROM users WHERE email = $1", req.Email).Scan(&passwordHash)
	require.NoError(t, err)

	// Проверяем, что хеш соответствует формату bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	require.NoError(t, err)
}

func TestAuth_Login(t *testing.T) {
	env, cleanup := SetupTestEnv(t)
	defer cleanup()

	// Создаем тестового пользователя
	user, _ := createTestUser(t, env)

	// Пробуем залогиниться с правильным паролем
	loginReq := models.LoginRequest{
		Email:    user.Email,
		Password: "password123",
	}

	resp, err := makeRequest(env, "POST", "/api/auth/login", loginReq, "")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Пробуем залогиниться с неправильным паролем
	loginReq.Password = "wrongpassword"
	resp, err = makeRequest(env, "POST", "/api/auth/login", loginReq, "")
	require.NoError(t, err)
	require.Equal(t, 401, resp.StatusCode)
}
