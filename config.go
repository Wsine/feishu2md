package main

type Config struct {
  Feishu    Feishu `json:"feishu"`
  Output    Output `json:"output"`
}

type Feishu struct {
  AppId     string `json:"app_id"`
  AppSecret string `json:"app_secret"`
}

type Output struct {
  ImageDir  string `json:"image_dir"`
}
