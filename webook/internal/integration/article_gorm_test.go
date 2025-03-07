package integration

// import (
// 	"bytes"
// 	"encoding/json"
// 	"go-basic/webook/internal/domain"
// 	"go-basic/webook/internal/integration/startup"
// 	"go-basic/webook/internal/repository/dao/article"
// 	ijwt "go-basic/webook/internal/web/jwt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"
// 	"gorm.io/gorm"
// )

// // ArticleTestSuite 是 Article 的单元测试套件
// type ArticleTestSuite struct {
// 	suite.Suite
// 	server *gin.Engine
// 	db     *gorm.DB
// }

// // 在所有测试执行前，会执行 SetupSuite
// func (s *ArticleTestSuite) SetupSuite() {
// 	s.server = gin.Default()
// 	// 设置用户 id 在redis中
// 	s.server.Use(func(ctx *gin.Context) {
// 		ctx.Set("claims", &ijwt.UserClaims{
// 			Uid: 123,
// 		})
// 	})
// 	// 初始化数据库
// 	s.db = startup.InitDB()
// 	artHdl := startup.InitArticleHandler()
// 	artHdl.RegisterRoutes(s.server)
// }

// // 在所有测试执行后，会执行 TearDownSuite，清理测试数据并让自增 ID 从 1 开始
// func (s *ArticleTestSuite) TearDownTest() {
// 	s.db.Exec("TRUNCATE TABLE articles")
// 	s.db.Exec("TRUNCATE TABLE published_articles")
// }

// func (s *ArticleTestSuite) TestEdit() {
// 	t := s.T()
// 	testCases := []struct {
// 		name   string
// 		before func(t *testing.T)
// 		after  func(t *testing.T)
// 		// 预期输入
// 		article Article

// 		wantCode int
// 		// 希望HTTP相应带上帖子的ID
// 		wantResult Result[int64]
// 	}{
// 		{
// 			name:   "新建帖子-保存成功",
// 			before: func(t *testing.T) {},
// 			after: func(t *testing.T) {
// 				// 验证数据库
// 				var art article.Article
// 				err := s.db.Where("id=?", "1").First(&art).Error
// 				assert.NoError(t, err)
// 				assert.True(t, art.Ctime > 0)
// 				assert.True(t, art.Utime > 0)
// 				art.Ctime = 0
// 				art.Utime = 0
// 				assert.Equal(t, article.Article{
// 					Id:       1,
// 					Title:    "标题",
// 					Content:  "内容",
// 					AuthorId: 123,
// 					Ctime:    0,
// 					Utime:    0,
// 					Status:   uint8(domain.ArticleStatusUnpublished),
// 				}, art)
// 			},
// 			article: Article{
// 				Title:   "标题",
// 				Content: "内容",
// 			},
// 			wantCode: http.StatusOK,
// 			wantResult: Result[int64]{
// 				Data: 1,
// 				Msg:  "OK",
// 			},
// 		},
// 		{
// 			name: "修改帖子-保存成功",
// 			before: func(t *testing.T) {
// 				err := s.db.Create(article.Article{
// 					Id:       2,
// 					Title:    "标题",
// 					Content:  "内容",
// 					AuthorId: 123,
// 					// 跟时间有关的测试，不要用 time.Now()，因为时间会变化
// 					Ctime:  123,
// 					Utime:  123,
// 					Status: uint8(domain.ArticleStatusPublished),
// 				}).Error
// 				assert.NoError(t, err)
// 			},
// 			after: func(t *testing.T) {
// 				// 验证数据库
// 				var art article.Article
// 				err := s.db.Where("id=?", "2").First(&art).Error
// 				assert.NoError(t, err)
// 				// 确保更新了 utime
// 				assert.True(t, art.Utime > 123)
// 				art.Utime = 0
// 				assert.Equal(t, article.Article{
// 					Id:       2,
// 					Title:    "新的标题",
// 					Content:  "新的内容",
// 					AuthorId: 123,
// 					Ctime:    123,
// 					Utime:    0,
// 					Status:   uint8(domain.ArticleStatusUnpublished),
// 				}, art)
// 			},
// 			article: Article{
// 				Id:      2,
// 				Title:   "新的标题",
// 				Content: "新的内容",
// 			},
// 			wantCode: http.StatusOK,
// 			wantResult: Result[int64]{
// 				Data: 2,
// 				Msg:  "OK",
// 			},
// 		},
// 		{
// 			name: "修改帖子-别人的帖子",
// 			before: func(t *testing.T) {
// 				err := s.db.Create(article.Article{
// 					Id:       3,
// 					Title:    "标题",
// 					Content:  "内容",
// 					AuthorId: 789,
// 					// 跟时间有关的测试，不要用 time.Now()，因为时间会变化
// 					Ctime:  123,
// 					Utime:  123,
// 					Status: uint8(domain.ArticleStatusPublished),
// 				}).Error
// 				assert.NoError(t, err)
// 			},
// 			after: func(t *testing.T) {
// 				// 验证数据库
// 				var art article.Article
// 				err := s.db.Where("id=?", "3").First(&art).Error
// 				assert.NoError(t, err)
// 				assert.Equal(t, article.Article{
// 					Id:       3,
// 					Title:    "标题",
// 					Content:  "内容",
// 					AuthorId: 789,
// 					Ctime:    123,
// 					Utime:    123,
// 					Status:   uint8(domain.ArticleStatusPublished),
// 				}, art)
// 			},
// 			article: Article{
// 				Id:      3,
// 				Title:   "新的标题",
// 				Content: "新的内容",
// 			},
// 			wantCode: http.StatusOK,
// 			wantResult: Result[int64]{
// 				Code: 5,
// 				Msg:  "系统错误",
// 			},
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// 构造请求
// 			// 执行
// 			// 验证数据
// 			tc.before(t)
// 			reqBody, err := json.Marshal(tc.article)
// 			assert.NoError(t, err)
// 			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(reqBody))
// 			require.NoError(t, err)
// 			req.Header.Set("Content-Type", "application/json")

// 			resp := httptest.NewRecorder()
// 			s.server.ServeHTTP(resp, req)
// 			assert.Equal(t, tc.wantCode, resp.Code)
// 			if resp.Code != http.StatusOK {
// 				return
// 			}
// 			var webRes Result[int64]
// 			err = json.NewDecoder(resp.Body).Decode(&webRes)
// 			require.NoError(t, err)
// 			assert.Equal(t, tc.wantResult, webRes)
// 			tc.after(t)
// 		})
// 	}
// }

// func (s *ArticleTestSuite) TestPublish() {
// 	t := s.T()
// 	testCases := []struct {
// 		name       string
// 		before     func(t *testing.T)
// 		after      func(t *testing.T)
// 		article    Article
// 		wantCode   int
// 		wantResult Result[int64]
// 	}{
// 		{
// 			name: "新建帖子并发表",
// 			before: func(t *testing.T) {
// 				// 什么也不需要做
// 			},
// 			after: func(t *testing.T) {
// 				// 验证一下数据
// 				var art article.Article
// 				err := s.db.Where("author_id = ?", 123).First(&art).Error
// 				assert.NoError(t, err)
// 				assert.True(t, art.Id > 0)
// 				assert.True(t, art.Ctime > 0)
// 				assert.True(t, art.Utime > 0)
// 				art.Ctime = 0
// 				art.Utime = 0
// 				art.Id = 0
// 				assert.Equal(t, article.Article{
// 					Title:    "hello，你好",
// 					Content:  "随便试试",
// 					AuthorId: 123,
// 					Status:   uint8(domain.ArticleStatusPublished),
// 				}, art)
// 				var publishedArt article.PublishedArticle
// 				err = s.db.Where("author_id = ?", 123).First(&publishedArt).Error
// 				assert.NoError(t, err)
// 				assert.True(t, publishedArt.Id > 0)
// 				assert.True(t, publishedArt.Ctime > 0)
// 				assert.True(t, publishedArt.Utime > 0)
// 				publishedArt.Ctime = 0
// 				publishedArt.Utime = 0
// 				publishedArt.Id = 0
// 				assert.Equal(t, article.PublishedArticle{
// 					Article: article.Article{
// 						Title:    "hello，你好",
// 						Content:  "随便试试",
// 						AuthorId: 123,
// 						Status:   uint8(domain.ArticleStatusPublished),
// 					},
// 				}, publishedArt)
// 			},
// 			article: Article{
// 				Title:   "hello，你好",
// 				Content: "随便试试",
// 			},
// 			wantCode: 200,
// 			wantResult: Result[int64]{
// 				Data: 1,
// 				Msg:  "OK",
// 			},
// 		},
// 		{
// 			// 制作库有，但是线上库没有
// 			name: "更新帖子并新发表",
// 			before: func(t *testing.T) {
// 				// 模拟已经存在的帖子
// 				err := s.db.Create(&article.Article{
// 					Id:       2,
// 					Title:    "我的标题",
// 					Content:  "我的内容",
// 					Ctime:    456,
// 					Status:   uint8(domain.ArticleStatusPublished),
// 					Utime:    234,
// 					AuthorId: 123,
// 				}).Error
// 				assert.NoError(t, err)
// 			},
// 			after: func(t *testing.T) {
// 				// 验证一下数据
// 				var art article.Article
// 				err := s.db.Where("id = ?", 2).First(&art).Error
// 				assert.NoError(t, err)
// 				assert.True(t, art.Utime > 234)
// 				art.Utime = 0
// 				assert.Equal(t, article.Article{
// 					Id:       2,
// 					Ctime:    456,
// 					Utime:    0,
// 					Content:  "新的内容",
// 					Title:    "新的标题",
// 					Status:   uint8(domain.ArticleStatusPublished),
// 					AuthorId: 123,
// 				}, art)
// 				var publishedArt article.PublishedArticle
// 				s.db.Where("id = ?", 2).First(&publishedArt)
// 				assert.True(t, publishedArt.Utime > 0)
// 				publishedArt.Ctime = 0
// 				publishedArt.Utime = 0
// 				assert.Equal(t, article.PublishedArticle{
// 					Article: article.Article{
// 						Id:       2,
// 						Title:    "新的标题",
// 						Content:  "新的内容",
// 						AuthorId: 123,
// 						Status:   uint8(domain.ArticleStatusPublished),
// 						Utime:    0,
// 						Ctime:    0,
// 					}},
// 					publishedArt)
// 			},
// 			article: Article{
// 				Id:      2,
// 				Title:   "新的标题",
// 				Content: "新的内容",
// 			},
// 			wantCode: 200,
// 			wantResult: Result[int64]{
// 				Data: 2,
// 				Msg:  "OK",
// 			},
// 		},
// 		{
// 			name: "更新帖子，并且重新发表",
// 			before: func(t *testing.T) {
// 				art := article.Article{
// 					Id:       3,
// 					Title:    "我的标题",
// 					Content:  "我的内容",
// 					Ctime:    456,
// 					Status:   1,
// 					Utime:    234,
// 					AuthorId: 123,
// 				}
// 				s.db.Create(&art)
// 				part := article.PublishedArticle{
// 					Article: art,
// 				}
// 				s.db.Create(&part)
// 			},
// 			after: func(t *testing.T) {
// 				var art article.Article
// 				s.db.Where("id = ?", 3).First(&art)
// 				assert.Equal(t, "新的标题", art.Title)
// 				assert.Equal(t, "新的内容", art.Content)
// 				assert.Equal(t, int64(123), art.AuthorId)
// 				assert.Equal(t, uint8(2), art.Status)
// 				// 创建时间没变
// 				assert.Equal(t, int64(456), art.Ctime)
// 				// 更新时间变了
// 				assert.True(t, art.Utime > 234)

// 				var part article.PublishedArticle
// 				s.db.Where("id = ?", 3).First(&part)
// 				assert.Equal(t, "新的标题", part.Title)
// 				assert.Equal(t, "新的内容", part.Content)
// 				assert.Equal(t, int64(123), part.AuthorId)
// 				assert.Equal(t, uint8(2), part.Status)
// 				// 创建时间没变
// 				assert.Equal(t, int64(456), part.Ctime)
// 				// 更新时间变了
// 				assert.True(t, part.Utime > 234)
// 			},
// 			article: Article{
// 				Id:      3,
// 				Title:   "新的标题",
// 				Content: "新的内容",
// 			},
// 			wantCode: 200,
// 			wantResult: Result[int64]{
// 				Data: 3,
// 				Msg:  "OK",
// 			},
// 		},
// 		{
// 			name: "更新别人的帖子，并且发表失败",
// 			before: func(t *testing.T) {
// 				art := article.Article{
// 					Id:      4,
// 					Title:   "我的标题",
// 					Content: "我的内容",
// 					Ctime:   456,
// 					Utime:   234,
// 					Status:  1,
// 					// 注意。这个 AuthorID 我们设置为另外一个人的ID
// 					AuthorId: 789,
// 				}
// 				s.db.Create(&art)
// 				part := article.PublishedArticle{
// 					Article: article.Article{
// 						Id:       4,
// 						Title:    "我的标题",
// 						Content:  "我的内容",
// 						Ctime:    456,
// 						Status:   2,
// 						Utime:    234,
// 						AuthorId: 789,
// 					}}
// 				s.db.Create(&part)
// 			},
// 			after: func(t *testing.T) {
// 				// 更新应该是失败了，数据没有发生变化
// 				var art article.Article
// 				s.db.Where("id = ?", 4).First(&art)
// 				assert.Equal(t, "我的标题", art.Title)
// 				assert.Equal(t, "我的内容", art.Content)
// 				assert.Equal(t, int64(456), art.Ctime)
// 				assert.Equal(t, int64(234), art.Utime)
// 				assert.Equal(t, uint8(1), art.Status)
// 				assert.Equal(t, int64(789), art.AuthorId)

// 				var part article.PublishedArticle
// 				// 数据没有变化
// 				s.db.Where("id = ?", 4).First(&part)
// 				assert.Equal(t, "我的标题", part.Title)
// 				assert.Equal(t, "我的内容", part.Content)
// 				assert.Equal(t, int64(789), part.AuthorId)
// 				assert.Equal(t, uint8(2), part.Status)
// 				// 创建时间没变
// 				assert.Equal(t, int64(456), part.Ctime)
// 				// 更新时间变了
// 				assert.Equal(t, int64(234), part.Utime)
// 			},
// 			article: Article{
// 				Id:      4,
// 				Title:   "新的标题",
// 				Content: "新的内容",
// 			},
// 			wantCode: 200,
// 			wantResult: Result[int64]{
// 				Code: 5,
// 				Msg:  "系统错误",
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.before(t)
// 			reqBody, err := json.Marshal(tc.article)
// 			assert.NoError(t, err)
// 			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer(reqBody))
// 			require.NoError(t, err)
// 			req.Header.Set("Content-Type", "application/json")
// 			resp := httptest.NewRecorder()
// 			s.server.ServeHTTP(resp, req)
// 			assert.Equal(t, tc.wantCode, resp.Code)
// 			if resp.Code != http.StatusOK {
// 				return
// 			}
// 			var webRes Result[int64]
// 			err = json.NewDecoder(resp.Body).Decode(&webRes)
// 			require.NoError(t, err)
// 			assert.Equal(t, tc.wantResult, webRes)
// 			tc.after(t)
// 		})
// 	}
// }

// func (s *ArticleTestSuite) TestABC() {
// 	s.T().Log("这个是测试套件")
// }

// func TestArticle(t *testing.T) {
// 	suite.Run(t, &ArticleTestSuite{})
// }

// type Article struct {
// 	Id      int64  `json:"id"`
// 	Title   string `json:"title"`
// 	Content string `json:"content"`
// }

// type Result[T any] struct {
// 	Code int    `json:"code"`
// 	Msg  string `json:"msg"`
// 	Data T      `json:"data"`
// }
