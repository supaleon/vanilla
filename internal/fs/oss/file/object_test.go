package file

var (
	testKeys = []string{
		".",
		"./",
		"../",
		"mytest.txt",
		"",
	}
	testObjects = []struct {
		key      string
		partSize int64
		isDir    bool
	}{
		{"hello.txt", 1024 * 1024 * 6, false},
		{"goodbye.txt", 0, false},
		{"level01/leve02/level03", 0, true},
		{"level08/leve020/level030", 0, true},
		{"level01/leve02/level03/day.txt", 0, false},
		{"level08/leve020/level030/nice.txt", 1024 * 1024 * 16, false},
	}
)
