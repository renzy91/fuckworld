package dao

func Init() {
	if err := InitItemView(); err != nil {
		panic(err)
	}
	if err := InitImageView(); err != nil {
		panic(err)
	}
}
