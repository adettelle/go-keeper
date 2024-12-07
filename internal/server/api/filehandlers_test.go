package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/mocks"
	"github.com/carlmjohnson/requests"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type readerDataStr struct {
	expected string
}

func readerContentsEq(s string) gomock.Matcher {
	return &readerDataStr{expected: s}
}

func (rd *readerDataStr) Matches(x interface{}) bool {
	if reader, ok := x.(io.Reader); !ok {
		return false
	} else {
		content, err := io.ReadAll(reader)
		return err == nil && bytes.Equal(content, []byte(rd.expected))
	}
}

func (rd *readerDataStr) String() string {
	return rd.expected
}

// ------- Хендлер: PUT /api/user/file
func TestFileAdd(t *testing.T) {
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)
	minioService := mocks.NewMockIMinioService(ctrl)

	h := &FileHandlers{
		FileRepo:     fileRepo,
		MinioService: minioService,
		JwtSignKey:   []byte("my_key"),
	}

	login := "Ane"
	userID := "123"

	file := fileCreateRequestDTO{
		FileName:    "./testdata/test.txt",
		Title:       "txt",
		Description: "text txt file",
	}
	stat, err := os.Stat(file.FileName)
	require.NoError(t, err)

	fileNameWithoutPath := filepath.Base(file.FileName)

	fileRepo.EXPECT().FileExists(gomock.Any(), file.Title, userID).Return(false, nil)

	minioService.EXPECT().Upload(gomock.Any(), readerContentsEq("hello!")).Return(nil)

	fileRepo.EXPECT().AddFile(gomock.Any(), fileNameWithoutPath, file.Title,
		file.Description, gomock.Any(), login).Return(nil)

	request, err := requests.
		URL("/api/user/file").
		Method(http.MethodPut).
		Header("x-file-name", fileNameWithoutPath).
		Header("x-file-title", file.Title).
		Header("x-file-description", file.Description).
		Header("Content-Length", strconv.Itoa(int(stat.Size()))).
		Header("x-user", login).
		Header("x-user-id", userID).
		BodyFile(file.FileName).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()

	h.FileAdd(response, request)

	require.Equal(t, wantHTTPStatus, response.Code)
}

func TestFileAddFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)
	minioService := mocks.NewMockIMinioService(ctrl)

	h := &FileHandlers{
		FileRepo:     fileRepo,
		MinioService: minioService,
		JwtSignKey:   []byte("my_key"),
	}

	login := "Ane"
	userID := "123"

	file := fileCreateRequestDTO{
		FileName:    "./testdata/test.txt",
		Title:       "txt",
		Description: "text txt file",
	}
	stat, err := os.Stat(file.FileName)
	require.NoError(t, err)

	fileNameWithoutPath := filepath.Base(file.FileName)

	fileRepo.EXPECT().FileExists(gomock.Any(), file.Title, userID).Return(false, nil)

	minioService.EXPECT().Upload(gomock.Any(), readerContentsEq("hello!")).Return(nil)

	fileRepo.EXPECT().AddFile(gomock.Any(), fileNameWithoutPath, file.Title,
		file.Description, gomock.Any(), login).Return(fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/file").
		Method(http.MethodPut).
		Header("x-file-name", fileNameWithoutPath).
		Header("x-file-title", file.Title).
		Header("x-file-description", file.Description).
		Header("Content-Length", strconv.Itoa(int(stat.Size()))).
		Header("x-user", login).
		Header("x-user-id", userID).
		BodyFile(file.FileName).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()

	h.FileAdd(response, request)

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: GET /api/user/files
func TestAllFiles(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &FileHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	files := []repo.FileToGet{
		{
			FileName:    "./file1.png",
			Title:       "png1",
			Description: "description1",
		},
		{
			FileName:    "./file2.png",
			Title:       "png2",
			Description: "description2",
		},
	}
	fileRepo.EXPECT().GetAllFiles(gomock.Any(), login).Return(files, nil)

	request, err := requests.
		URL("/api/user/files").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.AllFiles(response, request)

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, wantHTTPStatus, response.Code)

	expectedBody, err := json.Marshal([]fileGetRequestDTO{
		{
			FileName:    "./file1.png",
			Title:       "png1",
			Description: "description1",
		},
		{
			FileName:    "./file2.png",
			Title:       "png2",
			Description: "description2",
		},
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBody), string(resBody))
}

// ------- Хендлер: GET /api/user/files
func TestAllFilesFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &FileHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	fileRepo.EXPECT().GetAllFiles(gomock.Any(), login).Return(nil, fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/files").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.AllFiles(response, request)

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: GET /api/user/file/{title}
func TestFileByTitle(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)
	minioService := mocks.NewMockIMinioService(ctrl)

	// создаём объект-заглушку
	h := &FileHandlers{
		FileRepo:     fileRepo,
		MinioService: minioService,
		JwtSignKey:   []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	fileRepo.EXPECT().GetFileCoudIDByTitle(gomock.Any(), "title1", login).Return("cloudID123", nil)

	minioService.EXPECT().GetObject("cloudID123").Return(strings.NewReader("someFileContents"), nil)

	request, err := requests.
		URL("/api/user/file/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.FileGetByTitle(response, request)
	result := response.Result()
	defer result.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, resBody, []byte("someFileContents"))
	require.Equal(t, wantHTTPStatus, response.Code)
}

func TestFileByTitleInvalidTitle(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)
	minioService := mocks.NewMockIMinioService(ctrl)

	// создаём объект-заглушку
	h := &FileHandlers{
		FileRepo:     fileRepo,
		MinioService: minioService,
		JwtSignKey:   []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	fileRepo.EXPECT().GetFileCoudIDByTitle(gomock.Any(), "title1", login).Return("", nil)

	// minioService.EXPECT().GetObject("cloudID123").Return(strings.NewReader("someFileContents"), nil)

	request, err := requests.
		URL("/api/user/file/title1").
		Method(http.MethodGet).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusNotFound

	response := httptest.NewRecorder()

	h.FileGetByTitle(response, request)

	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Delete /api/user/file/{title}
func TestFIleDelete(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &FileHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	fileRepo.EXPECT().DeleteFileByTitle(gomock.Any(), "title1", login).Return(nil)

	request, err := requests.
		URL("/api/user/file/title1").
		Method(http.MethodDelete).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusOK

	response := httptest.NewRecorder()
	h.FileDeleteByTitle(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

// ------- Хендлер: Post /api/user/file/update/{title}
func TestFileUpdate(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &FileHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	title := "title1"
	fname := "file1.png"
	description := "png file"

	file := fileUpdateRequestDTO{
		FileName:    &fname,
		Description: &description,
	}

	fileRepo.EXPECT().UpdateFile(gomock.Any(), title, file.FileName, file.Description, userID).Return(nil)

	request, err := requests.
		URL("/api/user/file/update/"+title).
		Method(http.MethodPost).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		BodyJSON(&file).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusAccepted

	response := httptest.NewRecorder()
	h.FileUpdate(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}

func TestFileUpdateFailure(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)

	fileRepo := mocks.NewMockIFileRepo(ctrl)

	// создаём объект-заглушку
	h := &FileHandlers{
		FileRepo:   fileRepo,
		JwtSignKey: []byte("my_key"),
	}

	login := "Ane"
	userID := 123

	title := "title1"
	fname := "file1.png"
	description := "png file"

	file := fileUpdateRequestDTO{
		FileName:    &fname,
		Description: &description,
	}

	fileRepo.EXPECT().UpdateFile(gomock.Any(), title, file.FileName, file.Description,
		userID).Return(fmt.Errorf("DB error"))

	request, err := requests.
		URL("/api/user/file/update/"+title).
		Method(http.MethodPost).
		Header("x-user", login).
		Header("x-user-id", strconv.Itoa(userID)).
		BodyJSON(&file).
		Request(context.Background())
	require.NoError(t, err)

	request.SetPathValue("title", "title1")

	wantHTTPStatus := http.StatusInternalServerError

	response := httptest.NewRecorder()
	h.FileUpdate(response, request)
	require.Equal(t, wantHTTPStatus, response.Code)
}
