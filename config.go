package main

type Config struct {
	Postgres struct {
		Host     string
		Port     int
		User     string
		Password string
		DBSig    string
	}
}

// see https://www.4byte.directory/docs/
type ApiRes struct {
	Count    int64  `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		BytesSignature string `json:"bytes_signature"`
		CreatedAt      string `json:"created_at"`
		HexSignature   string `json:"hex_signature"`
		ID             int64  `json:"id"`
		TextSignature  string `json:"text_signature"`
	} `json:"results"`
}

type Resume struct {
	Url string `gorm:"type:text"`
}

type Signature struct {
	Code      []byte `gorm:"type:bytea"`
	Signature string `gorm:"type:text"`
}

func getConfig() Config {
	return Config{
		Postgres: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBSig    string
		}{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "your_password_here",
			DBSig:    "signatures_db",
		},
	}
}
