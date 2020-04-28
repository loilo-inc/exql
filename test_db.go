package exql

func testDb() DB {
	db, err := Open(&OpenOptions{
		Url: "root:@tcp(127.0.0.1:3326)/exql?charset=utf8mb4&parseTime=True&loc=Local",
	})
	if err != nil {
		panic(err)
	}
	return db
}
