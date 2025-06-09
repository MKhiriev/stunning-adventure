package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricHandler(t *testing.T) {
	h := initHandler()

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		route      string
		httpMethod string
		want       want
	}{
		{
			name:       "positive counter test #1",
			route:      "/update/counter/someMetric/527",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name:       "positive gauge test #2",
			route:      "/update/gauge/Alloc/9999999",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name:       "negative test #3 - no metric's name",
			route:      "/update/gauge/9999999",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain",
			},
		},
		{
			name:       "negative test #4 - http.Method is not POST",
			route:      "/update/gauge/Alloc/9999999",
			httpMethod: http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain",
			},
		},
		{
			name:       "negative test #5 - unknown type of metric",
			route:      "/update/otherType/Alloc/9999999",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
		{
			name:       "negative test #6 - wrong value type",
			route:      "/update/otherType/Alloc/wrongValue",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.httpMethod, test.route, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			h.MetricHandler(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func initHandler() *Handler {
	return NewHandler()
}
