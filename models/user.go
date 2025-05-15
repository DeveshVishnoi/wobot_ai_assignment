package models

type User struct {
	ID          string `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Password    string `json:"password" bson:"password"`
	Username    string `json:"username" bson:"username"`
	UsedStorage int64  `json:"used_storage" bson:"used_storage"`
	Quota       int64  `json:"quota" bson:"quota"`
	CreatedAt   int64  `json:"createdAt" bson:"createdAt"`
}

type UserContext struct {
	ID          string `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Username    string `json:"username" bson:"username"`
	UsedStorage int64  `json:"used_storage" bson:"used_storage"`
	Quota       int64  `json:"quota" bson:"quota"`
}

type UsernameAndPassword struct {
	Password string `json:"password" bson:"password"`
	Username string `json:"username" bson:"username"`
}

type UserSession struct {
	ID        string `json:"id" bson:"id"`
	UserID    string `json:"userId" bson:"userId"`
	StartTime int64  `json:"startTime" bson:"startTime"`
	EndTime   int64  `json:"endTime" bson:"endTime"`
	Token     string `json:"token" bson:"token"`
}
