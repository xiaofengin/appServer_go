package model

type AppUserInfo struct {
	ID                int    `json:"id" gorm:"type:bigint(20) auto_increment;primary_key"`
	UserAccount       string `json:"user_account"`
	UserPassword      string `json:"user_password"`
	AgoraChatUserName string `json:"agora_chat_user_name"`
	AgoraChatUserUuid string `json:"agora_chat_user_uuid"`
	AgoraUid          int    `json:"agora_uid" gorm:"type:bigint(20)"`
}
