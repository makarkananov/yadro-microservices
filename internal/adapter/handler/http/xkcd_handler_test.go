package http

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockComicService struct {
	mock.Mock
}

func (m *MockComicService) UpdateComics(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockComicService) GetNumberOfComics() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockComicService) Search(query string) ([]string, error) {
	args := m.Called(query)
	return args.Get(0).([]string), args.Error(1)
}

func TestUpdateComicsSuccess(t *testing.T) {
	service := new(MockComicService)
	service.On("GetNumberOfComics").Return(10).Once()
	service.On("UpdateComics", mock.Anything).Return(nil).Once()
	service.On("GetNumberOfComics").Return(15).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	rr := httptest.NewRecorder()
	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	service.AssertExpectations(t)
}

func TestUpdateComicsFailure(t *testing.T) {
	service := new(MockComicService)
	service.On("GetNumberOfComics").Return(10)
	service.On("UpdateComics", mock.Anything).Return(errors.New("update error")).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	rr := httptest.NewRecorder()
	handler.Update(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	service.AssertExpectations(t)
}

func TestSearchComicsSuccess(t *testing.T) {
	service := new(MockComicService)
	service.On("Search", "test").Return([]string{"url1", "url2"}, nil).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodGet, "/pics?search=test", nil)
	rr := httptest.NewRecorder()
	handler.Search(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	service.AssertExpectations(t)
}

func TestSearchComicsFailure(t *testing.T) {
	service := new(MockComicService)
	service.On("Search", "test").Return([]string{}, errors.New("search error")).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodGet, "/pics?search=test", nil)
	rr := httptest.NewRecorder()
	handler.Search(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	service.AssertExpectations(t)
}

func TestSearchComicsEmptyQuery(t *testing.T) {
	handler := NewXkcdHandler(nil)
	req, _ := http.NewRequest(http.MethodGet, "/pics?search=", nil)
	rr := httptest.NewRecorder()
	handler.Search(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
