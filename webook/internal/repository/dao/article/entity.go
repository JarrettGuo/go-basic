package article

type Article struct {
	//model
	Id int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	// 标题的长度
	// 正常都不会超过这个长度
	Title   string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 作者
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}

// PublishedArticle 衍生类型，偷个懒
type PublishedArticle Article

// PublishedArticleV1 s3 演示专属
type PublishedArticleV1 struct {
	Id       int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}

// type model struct {
// }

// func (u model) BeforeCreate(tx *gorm.DB) (err error) {
// 	startTime := time.Now()
// 	tx.Set("gorm:started_at", startTime)
// 	slog.Default().Info("这个是BeforeCreate")
// 	return nil
// }

// func (u model) AfterCreate(tx *gorm.DB) (err error) {
// 	val, _ := tx.Get("gorm:started_at")
// 	startTime, ok := val.(time.Time)
// 	if !ok {
// 		return nil
// 	}
// 	duration := time.Since(startTime)
// 	slog.Default().Info("这个是AfterCreate")
// 	return nil
// }
