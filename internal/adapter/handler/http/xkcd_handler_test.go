package http

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"yadro-microservices/internal/mocks"
)

func TestUpdateComicsSuccess(t *testing.T) {
	service := new(mocks.ComicService)
	service.On("GetNumberOfComics", mock.Anything).Return(10, nil).Once()
	service.On("UpdateComics", mock.Anything).Return(nil).Once()
	service.On("GetNumberOfComics", mock.Anything).Return(15, nil).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	rr := httptest.NewRecorder()
	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	service.AssertExpectations(t)
}

func TestUpdateComicsFailure(t *testing.T) {
	service := new(mocks.ComicService)
	service.On("GetNumberOfComics", mock.Anything).Return(10, nil).Once()
	service.On("UpdateComics", mock.Anything).Return(errors.New("update error")).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	rr := httptest.NewRecorder()
	handler.Update(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	service.AssertExpectations(t)
}

func TestUpdateComics_GetNumberOfComicsBeforeError(t *testing.T) {
	service := new(mocks.ComicService)
	service.On(
		"GetNumberOfComics",
		mock.Anything,
	).Return(0, errors.New("GetNumberOfComics error")).Twice()
	service.On("UpdateComics", mock.Anything).Return(nil).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	rr := httptest.NewRecorder()
	handler.Update(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	service.AssertExpectations(t)
}

func TestUpdateComics_GetNumberOfComicsAfterError(t *testing.T) {
	service := new(mocks.ComicService)
	service.On(
		"GetNumberOfComics",
		mock.Anything,
	).Return(0, errors.New("GetNumberOfComics error")).Twice()
	service.On("UpdateComics", mock.Anything).Return(nil).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	rr := httptest.NewRecorder()
	handler.Update(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	service.AssertExpectations(t)
}

func TestUpdateComics_EncodeError(t *testing.T) {
	service := new(mocks.ComicService)
	service.On("GetNumberOfComics", mock.Anything).Return(10, nil).Once()
	service.On("UpdateComics", mock.Anything).Return(nil).Once()
	service.On("GetNumberOfComics", mock.Anything).Return(15, nil).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	rr := httptest.NewRecorder()

	rr.Body = new(bytes.Buffer)
	rr.Body.Grow(1)

	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	service.AssertExpectations(t)
}

func TestSearchComicsSuccess(t *testing.T) {
	service := new(mocks.ComicService)
	service.On("Search", mock.Anything, "test").Return([]string{"url1", "url2"}, nil).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodGet, "/pics?search=test", nil)
	rr := httptest.NewRecorder()
	handler.Search(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	service.AssertExpectations(t)
}

func TestSearchComicsFailure(t *testing.T) {
	service := new(mocks.ComicService)
	service.On("Search", mock.Anything, "test").Return([]string{}, errors.New("search error")).Once()

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

func TestSearchComics_EncodeError(t *testing.T) {
	service := new(mocks.ComicService)
	service.On("Search", mock.Anything, "test").Return([]string{"url1", "url2"}, nil).Once()

	handler := NewXkcdHandler(service)
	req, _ := http.NewRequest(http.MethodGet, "/pics?search=test", nil)
	rr := httptest.NewRecorder()

	rr.Body = new(bytes.Buffer)
	rr.Body.Grow(1)

	handler.Search(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	service.AssertExpectations(t)
}
