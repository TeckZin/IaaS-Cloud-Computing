package models

type User struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Age        int32  `json:"age"`
	Department string `json:"department"`
}

type PostUser struct {
	Name       string `json:"name"`
	Age        int32  `json:"age"`
	Department string `json:"department"`
}
