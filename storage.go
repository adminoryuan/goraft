package main

var (
	config map[int]string
)

func init() {
	config = make(map[int]string)

	config[0] = "127.0.0.1:8880"

	config[1] = "127.0.0.1:8881"

	config[2] = "127.0.0.1:8882"

}
func getCurrAddr(id int) string {
	return config[id]
}
func getNodeLen() int {
	return len(config)
}
