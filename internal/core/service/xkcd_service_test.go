package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateComics(t *testing.T) {
	ctx := context.Background()

	clientMock := new(mocks.ComicClient)
	comicsRepMock := new(mocks.ComicRepository)
	processorMock := new(mocks.ComicProcessor)
	searchEngineMock := new(mocks.SearchEngine)

	service := NewXkcdService(clientMock, comicsRepMock, processorMock, searchEngineMock)

	existingIDs := map[int]bool{1: true, 2: true}
	newComics := domain.Comics{
		1: {Img: "https://example.com/comic3.png"},
	}

	comicsRepMock.On("GetAllIDs", mock.Anything).Return(existingIDs, nil)
	clientMock.On("GetComics", mock.Anything, existingIDs).Return(newComics, nil)
	comicsRepMock.On("Save", mock.Anything, newComics).Return(nil)
	searchEngineMock.On("CreateIndex", mock.Anything, newComics).Return(nil)

	err := service.UpdateComics(ctx)

	require.NoError(t, err)
	comicsRepMock.AssertExpectations(t)
	clientMock.AssertExpectations(t)
	comicsRepMock.AssertExpectations(t)
	searchEngineMock.AssertExpectations(t)
}

func TestUpdateComics_ErrorGettingIDs(t *testing.T) {
	ctx := context.Background()

	clientMock := new(mocks.ComicClient)
	comicsRepMock := new(mocks.ComicRepository)
	processorMock := new(mocks.ComicProcessor)
	searchEngineMock := new(mocks.SearchEngine)

	service := NewXkcdService(clientMock, comicsRepMock, processorMock, searchEngineMock)

	comicsRepMock.On(
		"GetAllIDs",
		mock.Anything,
	).Return(nil, errors.New("database error"))

	err := service.UpdateComics(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error extracting existing comic IDs")
	comicsRepMock.AssertExpectations(t)
	clientMock.AssertNotCalled(t, "GetComics", mock.Anything, mock.Anything)
	comicsRepMock.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	searchEngineMock.AssertNotCalled(t, "CreateIndex", mock.Anything, mock.Anything)
}

func TestSearch(t *testing.T) {
	ctx := context.Background()

	clientMock := new(mocks.ComicClient)
	comicsRepMock := new(mocks.ComicRepository)
	processorMock := new(mocks.ComicProcessor)
	searchEngineMock := new(mocks.SearchEngine)

	service := NewXkcdService(clientMock, comicsRepMock, processorMock, searchEngineMock)

	query := "test query"
	queryTokens := []string{"test", "query"}
	ids := []int{1, 2}
	comics := domain.Comics{
		1: {Img: "https://example.com/comic1.png"},
		2: {Img: "https://example.com/comic2.png"},
	}
	processorMock.On("FullProcess", query).Return(queryTokens, nil)
	searchEngineMock.On("Search", mock.Anything, queryTokens).Return(ids, nil)
	comicsRepMock.On("GetByID", ctx, 1).Return(comics[1], nil)
	comicsRepMock.On("GetByID", ctx, 2).Return(comics[2], nil)

	urls, err := service.Search(ctx, query)

	require.NoError(t, err)
	assert.Equal(t, []string{"https://example.com/comic1.png", "https://example.com/comic2.png"}, urls)
	processorMock.AssertExpectations(t)
	searchEngineMock.AssertExpectations(t)
	comicsRepMock.AssertExpectations(t)
}

func TestSearch_ErrorProcessingQuery(t *testing.T) {
	ctx := context.Background()

	clientMock := new(mocks.ComicClient)
	comicsRepMock := new(mocks.ComicRepository)
	processorMock := new(mocks.ComicProcessor)
	searchEngineMock := new(mocks.SearchEngine)

	service := NewXkcdService(clientMock, comicsRepMock, processorMock, searchEngineMock)

	query := "test query"
	processorMock.On(
		"FullProcess",
		query,
	).Return(nil, errors.New("processing error"))

	urls, err := service.Search(ctx, query)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error processing query")
	assert.Nil(t, urls)
	processorMock.AssertExpectations(t)
	searchEngineMock.AssertNotCalled(t, "Search", mock.Anything, mock.Anything)
	comicsRepMock.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

func TestScheduleUpdate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientMock := new(mocks.ComicClient)
	comicsRepMock := new(mocks.ComicRepository)
	processorMock := new(mocks.ComicProcessor)
	searchEngineMock := new(mocks.SearchEngine)

	updated := make(chan struct{})

	service := NewXkcdService(clientMock, comicsRepMock, processorMock, searchEngineMock)

	comicsRepMock.On("GetAllIDs", mock.Anything).Return(map[int]bool{}, nil)
	clientMock.On("GetComics", mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		updated <- struct{}{}
	}).Return(domain.Comics{}, nil)
	comicsRepMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	searchEngineMock.On("CreateIndex", mock.Anything, mock.Anything).Return(nil)

	updateTime := time.Now().Add(5 * time.Second)

	service.ScheduleUpdate(ctx, updateTime)

	select {
	case <-time.After(20 * time.Second):
	case <-updated:
	}

	comicsRepMock.AssertExpectations(t)
	clientMock.AssertExpectations(t)
	comicsRepMock.AssertExpectations(t)
	searchEngineMock.AssertExpectations(t)
}

func TestScheduleUpdate_CancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	clientMock := new(mocks.ComicClient)
	comicsRepMock := new(mocks.ComicRepository)
	processorMock := new(mocks.ComicProcessor)
	searchEngineMock := new(mocks.SearchEngine)

	updated := make(chan struct{})
	service := NewXkcdService(clientMock, comicsRepMock, processorMock, searchEngineMock)

	comicsRepMock.On("GetAllIDs", mock.Anything).Return(map[int]bool{}, nil)
	clientMock.On("GetComics", mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		updated <- struct{}{}
	}).Return(domain.Comics{}, nil)
	comicsRepMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	searchEngineMock.On("CreateIndex", mock.Anything, mock.Anything).Return(nil)

	updateTime := time.Now().Add(5 * time.Second)

	service.ScheduleUpdate(ctx, updateTime)

	cancel()

	select {
	case <-time.After(10 * time.Second):
	case <-updated:
		t.Errorf("update was not canceled")
	}

	clientMock.AssertNotCalled(t, "GetComics", mock.Anything, mock.Anything)
	comicsRepMock.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	searchEngineMock.AssertNotCalled(t, "CreateIndex", mock.Anything, mock.Anything)
}
