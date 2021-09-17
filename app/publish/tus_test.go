package publish

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lbryio/lbrytv/app/auth"
	"github.com/lbryio/lbrytv/app/sdkrouter"
	"github.com/lbryio/lbrytv/app/wallet"
	"github.com/lbryio/lbrytv/apps/lbrytv/config"
	"github.com/lbryio/lbrytv/internal/errors"
	"github.com/lbryio/lbrytv/internal/test"
	"github.com/lbryio/lbrytv/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
)

func newTusTestCfg(uploadPath string) tusd.Config {
	composer := tusd.NewStoreComposer()
	store := filestore.FileStore{
		Path: uploadPath,
	}
	store.UseIn(composer)
	return tusd.Config{
		BasePath:      "/api/v2/publish/",
		StoreComposer: composer,
	}
}

func newTestPublishHandler(t *testing.T) *PublishHandler {
	t.Helper()

	uploadPath := os.TempDir()

	h, err := NewHandler(
		newTestAuthProvider(),
		newTusTestCfg(uploadPath),
		uploadPath,
	)
	assert.Nil(t, err)
	return h
}

func newTestAuthProvider() auth.Provider {
	return auth.NewIAPIProvider(
		sdkrouter.New(config.GetLbrynetServers()),
		config.GetInternalAPIHost(),
	)
}

// tr is a helper to construct publish route path
var tr testRoute = "/api/v2/publish"

type testRoute string

func (tr testRoute) withID(id int) string {
	return fmt.Sprintf("%s/%d", tr, id)
}

func (tr testRoute) notify(id int) string {
	return fmt.Sprintf("%s/%d/notify", tr, id)
}

func (tr testRoute) root() string {
	return fmt.Sprintf("%s/", tr)
}

func newTestProvider(token, ip string) (*models.User, error) {
	reqChan := test.ReqChan()
	ts := test.MockHTTPServer(reqChan)
	var u *models.User
	if token == "uPldrToken" {
		u = &models.User{ID: 20404}
		u.R = u.R.NewStruct()
		u.R.LbrynetServer = &models.LbrynetServer{Address: ts.URL}
	}
	go func() {
		ts.NextResponse <- expectedStreamCreateResponse
	}()
	return u, nil
}

func newTestMux(h *PublishHandler, mwf ...mux.MiddlewareFunc) *mux.Router {
	router := mux.NewRouter()

	testRouter := router.PathPrefix("/api/v2/publish").Subrouter()
	for _, fn := range mwf {
		testRouter.Use(fn)
	}

	testRouter.HandleFunc("/", h.PostFile).Methods(http.MethodPost)
	testRouter.HandleFunc("/{id}", h.HeadFile).Methods(http.MethodHead)
	testRouter.HandleFunc("/{id}", h.PatchFile).Methods(http.MethodPatch)
	testRouter.HandleFunc("/{id}/notify", h.Notify).Methods(http.MethodPost)

	return testRouter
}

func newPartialUpload(t *testing.T, h *PublishHandler) string {
	t.Helper()

	testData := []byte("test file")
	r, err := http.NewRequest(http.MethodPost, tr.root(), bytes.NewReader(testData))
	assert.Nil(t, err)

	r.Header.Set(wallet.TokenHeader, "uPldrToken")
	r.Header.Set("Upload-Length", fmt.Sprintf("%d", len(testData)))

	w := httptest.NewRecorder()
	newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusCreated)
	return resp.Header.Get("location")
}

func newFinalUpload(t *testing.T, h *PublishHandler) string {
	loc := newPartialUpload(t, h)

	testData := []byte("test file")
	r, err := http.NewRequest(http.MethodPatch, loc, bytes.NewReader(testData))
	assert.Nil(t, err)

	r.Header.Set(wallet.TokenHeader, "uPldrToken")
	r.Header.Set("Content-Type", "application/offset+octet-stream")
	r.Header.Set("Upload-Offset", "0")

	w := httptest.NewRecorder()
	newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusNoContent)

	return loc
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	successTestCases := []struct {
		name string
		fn   func() (auth.Provider, tusd.Config, string)
	}{
		{
			name: "WithExistingDirectory",
			fn: func() (auth.Provider, tusd.Config, string) {
				uploadPath := os.TempDir()
				return newTestAuthProvider(), newTusTestCfg(uploadPath), uploadPath
			},
		},
		{
			name: "WithNewDirectory",
			fn: func() (auth.Provider, tusd.Config, string) {
				uploadPath := filepath.Join(os.TempDir(), "new_dir")
				return newTestAuthProvider(), newTusTestCfg(uploadPath), uploadPath
			},
		},
	}
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		for _, test := range successTestCases {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				h, err := NewHandler(test.fn())
				assert.Nil(t, err)
				assert.NotNil(t, h)
			})
		}
	})

	errorTestCases := []struct {
		name    string
		fn      func() (auth.Provider, tusd.Config, string)
		wantErr error
	}{
		{
			name: "WithNilAuthProvider",
			fn: func() (auth.Provider, tusd.Config, string) {
				uploadPath := os.TempDir()
				return nil, newTusTestCfg(uploadPath), uploadPath
			},
		},
		{
			name: "WithRestrictedDirectoryAccess",
			fn: func() (auth.Provider, tusd.Config, string) {
				if err := os.Mkdir("test_dir", 0444); err != nil { // read only
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.Remove("test_dir"); err != nil {
						t.Fatal(err)
					}
				})
				uploadPath := filepath.Join("test_dir", "new_dir")
				return newTestAuthProvider(), newTusTestCfg(uploadPath), uploadPath
			},
			wantErr: os.ErrPermission,
		},
	}
	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		for _, test := range errorTestCases {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				_, err := NewHandler(test.fn())
				assert.NotNil(t, err)

				if test.wantErr != nil {
					if err = errors.Unwrap(err); !errors.Is(err, test.wantErr) {
						t.Fatalf("expecting error: %+v, got: %+v", test.wantErr, err)
					}
				}
			})
		}
	})
}

func TestNotify(t *testing.T) {
	t.Parallel()

	t.Run("WithNoMiddleware", func(t *testing.T) {
		t.Parallel()

		h := newTestPublishHandler(t)
		w := httptest.NewRecorder()

		println(tr.notify(77))
		r, err := http.NewRequest(http.MethodPost, tr.notify(77), nil)
		assert.Nil(t, err)

		newTestMux(h).ServeHTTP(w, r)

		respBody, err := ioutil.ReadAll(w.Result().Body)
		assert.Nil(t, err)

		wantErrMsg := "auth.Middleware is required"
		gotErrMsg := test.StrToRes(t, string(respBody)).Error.Message
		assert.Equal(t, gotErrMsg, wantErrMsg)
	})

	t.Run("FileNotExist", func(t *testing.T) {
		t.Parallel()

		h := newTestPublishHandler(t)
		w := httptest.NewRecorder()

		r, err := http.NewRequest(http.MethodPost, tr.notify(404), nil)
		assert.Nil(t, err)

		r.Header.Set(wallet.TokenHeader, "uPldrToken")

		newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

		respBody, err := ioutil.ReadAll(w.Result().Body)
		assert.Nil(t, err)

		wantErrMsg := "no such file or directory"
		gotErrMsg := test.StrToRes(t, string(respBody)).Error.Message
		assert.Contains(t, gotErrMsg, wantErrMsg)
	})

	t.Run("UploadInProgress", func(t *testing.T) {
		t.Parallel()

		h := newTestPublishHandler(t)
		w := httptest.NewRecorder()

		loc := newPartialUpload(t, h)

		r, err := http.NewRequest(http.MethodPost, loc+"/notify", http.NoBody)
		assert.Nil(t, err)

		r.Header.Set(wallet.TokenHeader, "uPldrToken")

		newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

		respBody, err := ioutil.ReadAll(w.Result().Body)
		assert.Nil(t, err)

		wantErrMsg := "upload is still in process"
		gotErrMsg := test.StrToRes(t, string(respBody)).Error.Message
		assert.Equal(t, gotErrMsg, wantErrMsg)
	})

	t.Run("UploadCompleted", func(t *testing.T) {
		t.Parallel()

		h := newTestPublishHandler(t)
		w := httptest.NewRecorder()

		loc := newFinalUpload(t, h)

		r, err := http.NewRequest(
			http.MethodPost, loc+"/notify",
			strings.NewReader(expectedStreamCreateRequest),
		)
		assert.Nil(t, err)

		r.Header.Set(wallet.TokenHeader, "uPldrToken")

		newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

		response := w.Result()
		respBody, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)

		test.AssertEqualJSON(t, expectedStreamCreateResponse, respBody)
	})
}

func TestTus(t *testing.T) {
	t.Parallel()

	t.Run("CreateUpload", func(t *testing.T) {
		t.Parallel()

		h := newTestPublishHandler(t)
		w := httptest.NewRecorder()

		testData := []byte("test file")
		r, err := http.NewRequest(http.MethodPost, tr.root(), bytes.NewReader(testData))
		assert.Nil(t, err)

		r.Header.Set(wallet.TokenHeader, "uPldrToken")
		r.Header.Set("Upload-Length", fmt.Sprintf("%d", len(testData)))

		newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusCreated)
		assert.NotEmpty(t, resp.Header.Get("location"))
	})

	t.Run("ResumeUpload", func(t *testing.T) {
		t.Parallel()

		h := newTestPublishHandler(t)
		w := httptest.NewRecorder()

		loc := newPartialUpload(t, h)

		testData := []byte("test file")
		r, err := http.NewRequest(http.MethodPatch, loc, bytes.NewReader(testData))
		assert.Nil(t, err)

		r.Header.Set(wallet.TokenHeader, "uPldrToken")
		r.Header.Set("Content-Type", "application/offset+octet-stream")
		r.Header.Set("Upload-Offset", "0")

		newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusNoContent)
	})

	t.Run("QueryFile", func(t *testing.T) {
		t.Parallel()

		h := newTestPublishHandler(t)
		w := httptest.NewRecorder()

		loc := newPartialUpload(t, h)

		r, err := http.NewRequest(http.MethodHead, loc, http.NoBody)
		assert.Nil(t, err)

		r.Header.Set(wallet.TokenHeader, "uPldrToken")

		newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("QueryNonExistentFile", func(t *testing.T) {
		t.Parallel()

		h := newTestPublishHandler(t)
		w := httptest.NewRecorder()
		_ = newPartialUpload(t, h)

		r, err := http.NewRequest(http.MethodHead, tr.withID(404), http.NoBody)
		assert.Nil(t, err)

		r.Header.Set(wallet.TokenHeader, "uPldrToken")

		newTestMux(h, auth.Middleware(newTestProvider)).ServeHTTP(w, r)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusNotFound)
	})
}
