package service

import (
	"context"
	"go-basic/webook/internal/domain"
	"testing"
	"time"

	svcmocks "go-basic/webook/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBatchRankingService_topN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (ArticleService, InteractiveService)

		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "计算成功",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				// 最简单，一批搞完
				artSvc.EXPECT().ListPub(gomock.Any(), 0, 3, gomock.Any()).Return([]domain.Article{
					{Id: 1, Utime: now, Ctime: now},
					{Id: 2, Utime: now, Ctime: now},
					{Id: 3, Utime: now, Ctime: now},
				}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), 3, 3, gomock.Any()).Return([]domain.Article{}, nil)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3}).Return(map[int64]domain.Interactive{
					1: {BizId: 1, LikeCnt: 1},
					2: {BizId: 2, LikeCnt: 2},
					3: {BizId: 3, LikeCnt: 3}}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{}).Return(map[int64]domain.Interactive{}, nil)
				return artSvc, intrSvc
			},
			wantArts: []domain.Article{
				{Id: 3, Utime: now, Ctime: now},
				{Id: 2, Utime: now, Ctime: now},
				{Id: 1, Utime: now, Ctime: now},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc, intrSvc := tc.mock(ctrl)
			svc := &BatchRankingService{
				intrSvc:   intrSvc,
				artSvc:    artSvc,
				batchSize: 3,
				n:         3,
				scoreFunc: func(utime time.Time, likeCnt int64) float64 {
					return float64(likeCnt)
				},
			}
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
