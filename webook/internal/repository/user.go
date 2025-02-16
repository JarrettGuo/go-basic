package repository

type UserRepository struct{}

func (r *UserRepository) FindById(int64) {
	// 先从 cache 中查找
	// 如果 cache 中没有，再从数据库中查找
	// 找到后，将数据写入 cache
}
