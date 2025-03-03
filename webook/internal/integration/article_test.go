package integration

import (
	"bytes"
	"encoding/json"
	"go-basic/webook/internal/integration/startup"
	"go-basic/webook/internal/repository/dao"
	ijwt "go-basic/webook/internal/web/jwt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ArticleTestSuite 是 Article 的单元测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

// 在所有测试执行前，会执行 SetupSuite
func (s *ArticleTestSuite) SetupSuite() {
	s.server = gin.Default()
	// 设置用户 id 在redis中
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})
	// 初始化数据库
	s.db = startup.InitDB()
	artHdl := startup.InitArticleHandler()
	artHdl.RegisterRoutes(s.server)
}

// 在所有测试执行后，会执行 TearDownSuite，清理测试数据并让自增 ID 从 1 开始
func (s *ArticleTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE articles")
}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		// 预期输入
		article Article

		wantCode int
		// 希望HTTP相应带上帖子的ID
		wantResult Result[int64]
	}{
		{
			name:   "新建帖子-保存成功",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				// 验证数据库
				var art dao.Article
				err := s.db.Where("id=?", "1").First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.Equal(t, "标题", art.Title)
				assert.Equal(t, "内容", art.Content)
			},
			article: Article{
				Title:   "标题",
				Content: "内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 1,
				Msg:  "OK",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 构造请求
			// 执行
			// 验证数据
			tc.before(t)
			reqBody, err := json.Marshal(tc.article)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantResult, webRes)
			tc.after(t)
		})
	}
}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("这个是测试套件")
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
