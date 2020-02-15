package models

type ImgurUploadResponse struct {
	Data    ImgurUploadResponseData `json:"data"`
	Success bool                    `json:"success"`
}

type ImgurUploadResponseData struct {
	Link string `json:"link"`
}
