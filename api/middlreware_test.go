package api

import (
	"fmt"
	"net/http"
	"net/http/httptest" // Package for testing HTTP servers and handlers
	"testing"
	"time"

	"github.com/Oliver-Zen/simplebank/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// addAuthorization creates a new access token and add it to the authorization header of the [req]
func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	
	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	// set the healder of the request
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	// there are multiple target test cases, so use table-driven test
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(
					t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recoder.Code)
			},
		},
		{
			name: "NusupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(
					t,
					request,
					tokenMaker,
					"unsupported",
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recoder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(
					t,
					request,
					tokenMaker,
					"",
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recoder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(
					t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					"user",
					-time.Minute,
				)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recoder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)

			// add API route and handler just for test
			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker), // middleware
				func(ctx *gin.Context) { // handler func
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			// send request to the added API
			recorder := httptest.NewRecorder()                             // 创建一个模拟的 HTTP 响应记录器
			request, err := http.NewRequest(http.MethodGet, authPath, nil) // 模拟一个 GET 请求
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker) // 调用测试用例中的 setupAuth 函数，为 req 设置认证信息
			server.router.ServeHTTP(recorder, request)  // 将模拟的 req 传递给路由器处理，结果写入 recorder

			tc.checkResponse(t, recorder) // 检查路由器返回的 res 是否符合预期
		})
	}
}
