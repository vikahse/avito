package e2e_test

import (
	"avito/controllers"
	"avito/database"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GerBannerSuite struct {
	suite.Suite
}

func TestGetBannerSuite(t *testing.T) {
	suite.Run(t, new(GerBannerSuite))
}

func (s *GerBannerSuite) SetupTest() {
	config := map[string]string{
		"DB_HOST":     "127.0.0.1",
		"DB_DRIVER":   "postgres",
		"DB_USER":     "postgres",
		"DB_PASSWORD": "1234",
		"DB_NAME":     "service_db",
		"DB_PORT":     "5434",
		"REDIS_HOST":  "127.0.0.1:6379",
	}

	database.InitDatabase(config)

	database.GlobalDB.Exec("INSERT INTO tags (id, value) VALUES (100, 50), (101, 51), (102, 52)")
	database.GlobalDB.Exec(`INSERT INTO banners (id, feature_id, content, is_active) VALUES (200, 99, '{"test1": "test2"}', true)`)
	database.GlobalDB.Exec(`INSERT INTO banners (id, feature_id, content, is_active) VALUES (201, 98, '{"test2": "test2"}', false)`)
	database.GlobalDB.Exec(`INSERT INTO banner_tags (banner_id, tag_id) VALUES (200, 100), (200, 101), (201, 101)`)

	database.GlobalDB.Exec(`INSERT INTO users (id, name, password, is_admin) VALUES (1000, 'admin', '$2a$14$9T6hL2LrTCWRWJJ6uYObC.s9RjejdTOiyCNVCUvKeOCasAtc7ZLBq', true), (1001, 'user', '$2a$14$9T6hL2LrTCWRWJJ6uYObC.s9RjejdTOiyCNVCUvKeOCasAtc7ZLBq', false)`)
}

func (s *GerBannerSuite) TearDownTest() {
	database.GlobalDB.Exec("DELETE FROM banner_tags WHERE banner_id in (200, 201)")
	database.GlobalDB.Exec("DELETE FROM tags WHERE id IN (100, 101, 102)")
	database.GlobalDB.Exec("DELETE FROM banners WHERE id in (200, 201)")

	database.GlobalDB.Exec("DELETE FROM users WHERE id in (1000, 1001)")
}

func (s *GerBannerSuite) TestAdminBanner() {

	// Login
	loginPayload := map[string]string{
		"name":     "admin",
		"password": "avito2",
	}

	jsonData, err := json.Marshal(loginPayload)
	s.NoError(err)

	loginResp, err := http.Post("http://127.0.0.1:8008/login", "application/json", bytes.NewBuffer(jsonData))
	s.NoError(err)

	defer loginResp.Body.Close()

	s.Equal(http.StatusOK, loginResp.StatusCode)

	body, err := io.ReadAll(loginResp.Body)
	s.NoError(err)

	var loginTokens controllers.LoginResponse
	err = json.Unmarshal(body, &loginTokens)
	s.NoError(err)

	// Get banner
	payload := map[string]interface{}{
		"tag_id":           50,
		"feature_id":       99,
		"use_last_version": true,
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("GET", "http://127.0.0.1:8008/user_banner", bytes.NewBuffer(jsonPayload))
	s.NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", loginTokens.Token)

	client := &http.Client{}
	bannerResp, err := client.Do(req)
	s.NoError(err)

	defer bannerResp.Body.Close()

	s.Equal(http.StatusOK, bannerResp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(bannerResp.Body).Decode(&responseBody)
	s.NoError(err)

	s.Equal(float64(http.StatusOK), responseBody["Code"])
	s.Equal(map[string]interface{}(map[string]interface{}{"test1": "test2"}), responseBody["JSON-отображение баннера"])
}

func (s *GerBannerSuite) TestAuthNoAdmin() {
	// Login
	loginPayload := map[string]string{
		"name":     "user",
		"password": "avito2",
	}

	jsonData, err := json.Marshal(loginPayload)
	s.NoError(err)

	loginResp, err := http.Post("http://127.0.0.1:8008/login", "application/json", bytes.NewBuffer(jsonData))
	s.NoError(err)

	defer loginResp.Body.Close()

	s.Equal(http.StatusOK, loginResp.StatusCode)

	body, err := io.ReadAll(loginResp.Body)
	s.NoError(err)

	var loginTokens controllers.LoginResponse
	err = json.Unmarshal(body, &loginTokens)
	s.NoError(err)

	// Get banner
	payload := map[string]interface{}{
		"tag_id":           51,
		"feature_id":       98,
		"use_last_version": true,
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("GET", "http://127.0.0.1:8008/user_banner", bytes.NewBuffer(jsonPayload))
	s.NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", loginTokens.Token)

	client := &http.Client{}
	bannerResp, err := client.Do(req)
	s.NoError(err)

	defer bannerResp.Body.Close()

	s.Equal(http.StatusNotFound, bannerResp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(bannerResp.Body).Decode(&responseBody)
	s.NoError(err)

	s.Equal(float64(http.StatusNotFound), responseBody["Code"])
	s.Equal("Баннер не найден", responseBody["Error"])
}

func (s *GerBannerSuite) TestCache() {

	// Login
	loginPayload := map[string]string{
		"name":     "admin",
		"password": "avito2",
	}

	jsonData, err := json.Marshal(loginPayload)
	s.NoError(err)

	loginResp, err := http.Post("http://127.0.0.1:8008/login", "application/json", bytes.NewBuffer(jsonData))
	s.NoError(err)

	defer loginResp.Body.Close()

	s.Equal(http.StatusOK, loginResp.StatusCode)

	body, err := io.ReadAll(loginResp.Body)
	s.NoError(err)

	var loginTokens controllers.LoginResponse
	err = json.Unmarshal(body, &loginTokens)
	s.NoError(err)

	// Get banner
	payload := map[string]interface{}{
		"tag_id":           50,
		"feature_id":       99,
		"use_last_version": false,
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("GET", "http://127.0.0.1:8008/user_banner", bytes.NewBuffer(jsonPayload))
	s.NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", loginTokens.Token)

	client := &http.Client{}
	bannerResp, err := client.Do(req)
	s.NoError(err)

	defer bannerResp.Body.Close()

	s.Equal(http.StatusOK, bannerResp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(bannerResp.Body).Decode(&responseBody)
	s.NoError(err)

	s.Equal(float64(http.StatusOK), responseBody["Code"])
	s.Equal(map[string]interface{}(map[string]interface{}{"test1": "test2"}), responseBody["JSON-отображение баннера"])

	database.GlobalDB.Exec(`UPDATE banners SET content = '{"aaaa":"bbbb"}' WHERE id = 200`)

	// Get banner
	payload = map[string]interface{}{
		"tag_id":           50,
		"feature_id":       99,
		"use_last_version": false,
	}

	jsonPayload, _ = json.Marshal(payload)

	req, err = http.NewRequest("GET", "http://127.0.0.1:8008/user_banner", bytes.NewBuffer(jsonPayload))
	s.NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", loginTokens.Token)

	client = &http.Client{}
	bannerResp, err = client.Do(req)
	s.NoError(err)

	defer bannerResp.Body.Close()

	s.Equal(http.StatusOK, bannerResp.StatusCode)

	var responseBodySecond map[string]interface{}
	err = json.NewDecoder(bannerResp.Body).Decode(&responseBodySecond)
	s.NoError(err)

	s.Equal(float64(http.StatusOK), responseBodySecond["Code"])
	s.Equal(map[string]interface{}(map[string]interface{}{"test1": "test2"}), responseBodySecond["JSON-отображение баннера"])

	// Get banner
	payload = map[string]interface{}{
		"tag_id":           50,
		"feature_id":       99,
		"use_last_version": true,
	}

	jsonPayload, _ = json.Marshal(payload)

	req, err = http.NewRequest("GET", "http://127.0.0.1:8008/user_banner", bytes.NewBuffer(jsonPayload))
	s.NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", loginTokens.Token)

	client = &http.Client{}
	bannerResp, err = client.Do(req)
	s.NoError(err)

	defer bannerResp.Body.Close()

	s.Equal(http.StatusOK, bannerResp.StatusCode)

	var responseBodyThird map[string]interface{}
	err = json.NewDecoder(bannerResp.Body).Decode(&responseBodyThird)
	s.NoError(err)

	s.Equal(float64(http.StatusOK), responseBodyThird["Code"])
	s.Equal(map[string]interface{}(map[string]interface{}{"aaaa": "bbbb"}), responseBodyThird["JSON-отображение баннера"])
}
