package util

import (
	"net/http"
	"net/url"
	"math/rand"
)

// TokenAPI 获取带 token 的 API 地址
func TokenAPI(api, token string) (string, error) {
	u, err := url.Parse(api)
	if err != nil {
		return "", nil
	}
	query := u.Query()
	query.Set("access_token", token)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

// GetQuery returns url query value
func GetQuery(req *http.Request, key string) string {
	if values, ok := req.URL.Query()[key]; ok && len(values) > 0 {
		return values[0]
	}

	return ""
}

func NonceStr(length int) string {
	base := []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	baseLength := len(base)

	var nonceStrArr []byte
	for idx := 0; idx < length; idx++ {
		number := rand.Intn(baseLength)
		nonceStrArr = append(nonceStrArr, base[number])
	}

	return string(nonceStrArr)
}
