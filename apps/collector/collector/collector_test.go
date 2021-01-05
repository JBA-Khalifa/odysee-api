package collector

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/lbryio/lbrytv/apps/collector/models"
	"github.com/lbryio/lbrytv/config"
	"github.com/lbryio/lbrytv/internal/storage"
	"github.com/lbryio/lbrytv/pkg/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	appConfig := config.ReadConfig("collector")
	dbConfig := appConfig.GetDatabase()
	params := storage.ConnParams{
		Connection:     dbConfig.Connection,
		DBName:         dbConfig.DBName,
		Options:        dbConfig.Options,
		MigrationsPath: "../migrations",
	}
	dbConn, connCleanup := storage.CreateTestConn(params)
	dbConn.SetDefaultConnection()

	code := m.Run()

	connCleanup()
	os.Exit(code)
}

func TestHealthz(t *testing.T) {
	app := app.New("127.0.0.1:11111")
	app.InstallRoutes(RouteInstaller)
	app.Start()
	defer app.Shutdown()
	time.Sleep(200 * time.Millisecond)

	r, err := http.Get("http://127.0.0.1:11111/healthz")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.StatusCode)
}

func TestCORS(t *testing.T) {
	app := app.New("127.0.0.1:11111", app.AllowOrigin("*"))
	app.InstallRoutes(RouteInstaller)
	app.Start()
	defer app.Shutdown()
	time.Sleep(200 * time.Millisecond)

	req, err := http.NewRequest(http.MethodOptions, "http://127.0.0.1:11111/api/v1/events/video", nil)
	require.NoError(t, err)
	r, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	h := r.Header
	require.Equal(t, http.StatusOK, r.StatusCode)
	assert.Equal(t, "7200", h.Get("Access-Control-Max-Age"))
	assert.Equal(t, "*", h.Get("Access-Control-Allow-Origin"))
	assert.Equal(
		t,
		"Origin, X-Requested-With, Content-Type, Accept",
		h.Get("Access-Control-Allow-Headers"),
	)

	req, err = http.NewRequest(http.MethodPost, "http://127.0.0.1:11111/api/v1/events/video", nil)
	require.NoError(t, err)
	r, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	h = r.Header
	require.Equal(t, http.StatusBadRequest, r.StatusCode)
	assert.Equal(t, "7200", h.Get("Access-Control-Max-Age"))
	assert.Equal(t, "*", h.Get("Access-Control-Allow-Origin"))
	assert.Equal(
		t,
		"Origin, X-Requested-With, Content-Type, Accept",
		h.Get("Access-Control-Allow-Headers"),
	)
}

func TestEventHandler(t *testing.T) {
	type testData struct {
		name           string
		input          []byte
		expectedStatus int
		expectedBody   []byte
	}
	var tests []testData

	data := BufferingPost{
		Client: "aaa",
		Type:   "buffering",
		Data: BufferingPostData{
			URL:      "lbry://one",
			Position: 11654,
		},
	}
	serialized, err := json.Marshal(data)
	require.NoError(t, err)
	tests = append(tests, testData{"buffering event", serialized, http.StatusOK, []byte(``)})

	data = BufferingPost{
		Client: "aaa",
		Type:   "buffering",
		Device: "android",
		Data: BufferingPostData{
			URL:      "lbry://one",
			Position: 11654,
		},
	}
	serialized, err = json.Marshal(data)
	require.NoError(t, err)
	tests = append(tests, testData{"buffering event with device", serialized, http.StatusOK, []byte(``)})

	data = BufferingPost{
		Client: "aaa",
		Type:   "buffering",
		Device: "android",
		Data: BufferingPostData{
			Position: 11654,
		},
	}
	serialized, err = json.Marshal(data)
	require.NoError(t, err)
	tests = append(tests, testData{"buffering event with missing fields", serialized, http.StatusBadRequest, []byte(`Error at "/data":Doesn't match schema "oneOf"`)})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/events/video", bytes.NewReader(test.input))
			req.Header.Add("content-type", "application/json; charset=utf-8")
			req.Header.Add("host", "collector-service.dev.lbry.tv")
			rr := httptest.NewRecorder()
			EventHandler(rr, req)
			response := rr.Result()
			respBody, err := ioutil.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, test.expectedStatus, response.StatusCode)
			if test.expectedStatus != 200 {
				assert.Contains(t, string(respBody), string(test.expectedBody))
			}
		})
	}

	count, err := models.BufferEvents().CountG()
	require.NoError(t, err)
	assert.EqualValues(t, 2, count)
	e, err := models.BufferEvents(models.BufferEventWhere.ID.EQ(1)).OneG()
	require.NoError(t, err)
	assert.Equal(t, "aaa", e.Client)
	assert.Equal(t, 11654, e.Position)
	assert.Equal(t, "lbry://one", e.URL)
}
