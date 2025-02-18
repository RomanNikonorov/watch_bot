package watch

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockHTTPClient реализует интерфейс HTTPClient для мокирования
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return m.DoFunc(req)
}

// TestIsUrlOkHealthy проверяет, что метод IsUrlOk возвращает true для корректного URL
func TestIsUrlOkHealthy(t *testing.T) {
	// Создаем фейковый HTTP-клиент, который отвечает 200 OK
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("OK")),
			}, nil
		},
	}

	// Создаем экземпляр RealURLChecker
	checker := RealURLChecker{}

	// Вызываем метод IsUrlOk
	url := "http://example.com"
	unhealthyThreshold := 1
	unhealthyDelay := 1
	result := checker.IsUrlOk(url, unhealthyThreshold, unhealthyDelay, mockClient)

	// Проверяем результат
	if !result {
		t.Errorf("Expected IsUrlOk to return true, but got false")
	}
}

// TestIsUrlOkUnhealthy проверяет, что метод IsUrlOk возвращает false для некорректного URL
func TestIsUrlOkUnhealthy(t *testing.T) {
	// Создаем фейковый HTTP-клиент, который отвечает 404 Not Found
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("OK")),
			}, nil
		},
	}

	// Создаем экземпляр RealURLChecker
	checker := RealURLChecker{}

	// Вызываем метод IsUrlOk
	url := "http://example.com"
	unhealthyThreshold := 1
	unhealthyDelay := 1
	result := checker.IsUrlOk(url, unhealthyThreshold, unhealthyDelay, mockClient)

	// Проверяем результат
	if result {
		t.Errorf("Expected IsUrlOk to return false, but got true")
	}
}

// TestIsUrlOkUnreachable проверяет, что метод IsUrlOk возвращает false для недоступного URL
func TestIsUrlOkUnreachable(t *testing.T) {
	// Создаем фейковый HTTP-клиент, который всегда возвращает ошибку
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, http.ErrNoLocation
		},
	}

	// Создаем экземпляр RealURLChecker
	checker := RealURLChecker{}

	// Вызываем метод IsUrlOk
	url := "http://example.com"
	unhealthyThreshold := 1
	unhealthyDelay := 1
	result := checker.IsUrlOk(url, unhealthyThreshold, unhealthyDelay, mockClient)

	// Проверяем результат
	if result {
		t.Errorf("Expected IsUrlOk to return false, but got true")
	}
}
