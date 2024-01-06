package sendMessage

import (
	"net/http"
	"net/url"
	"strings"
)

func SendLineMessage(message string, token string) error {

	lineUrl := "https://notify-api.line.me/api/notify"
	const maxMessageLength = 800 //"더보기" 안나오게 끊기
	for len(message) > 0 {
		chunk := message
		if len(message) > maxMessageLength {
			chunk = message[:maxMessageLength]
			//중간에 짤리면 한글 깨짐. 마지막 공백을 찾아서 메세지 분리
			lastSpace := strings.LastIndex(chunk, " ")
			if lastSpace == -1 || lastSpace == 0 {
				message = message[maxMessageLength:]
			} else {

				chunk = message[:lastSpace]
				message = message[lastSpace+1:]
			}
		} else {
			chunk = message
			message = ""
		}

		data := url.Values{
			"message": {chunk},
		}

		req, err := http.NewRequest("POST", lineUrl, strings.NewReader(data.Encode()))
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	return nil
}
