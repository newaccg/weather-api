package model

type weatherNum float32

type WeatherResponse struct{
	ResolvedAddress string `json:"resolvedAddress"`
	Description string `json:"description"`
	Days []Day `json:"days"`
}

type Day struct{
	Conditions string `json:"conditions"`
	Datetime string `json:"datetime"`
	Temp weatherNum `json:"temp"`
	FeelsLike weatherNum `json:"feelslike"`
	Humidity weatherNum `json:"humidity"`
	Icon string `json:"icon"`
	Windspeed weatherNum `json:"windspeed"`
}
