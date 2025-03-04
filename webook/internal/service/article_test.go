package service

import (
	"context"
	"errors"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/article"
	repomocks "go-basic/webook/internal/repository/article/mocks"
	"go-basic/webook/pkg/logger"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository)
		art     domain.Article
		wantErr error
		wantId  int64
	}{
		{
			name: "新建发表成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "标题",
					Content: "内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "标题",
					Content: "内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				return author, reader
			},
			art: domain.Article{
				Title:   "标题",
				Content: "内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 1,
		},
		{
			name: "修改并发表成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				return author, reader
			},
			art: domain.Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 2,
		},
		{
			name: "保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(errors.New("mock db 错误"))
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				return author, reader
			},
			art: domain.Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  0,
			wantErr: errors.New("mock db 错误"),
		},
		{
			name: "保存到制作库成功，重试到线上库成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(errors.New("mock db 错误"))
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				return author, reader
			},
			art: domain.Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  2,
			wantErr: nil,
		},
		{
			name: "保存到制作库成功，重试到线上库失败",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Times(3).Return(errors.New("mock db 错误"))
				return author, reader
			},
			art: domain.Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  2,
			wantErr: errors.New("mock db 错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			author, reader := tc.mock(ctrl)
			svc := NewArticleServiceV1(author, reader, &logger.NopLogger{})
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
