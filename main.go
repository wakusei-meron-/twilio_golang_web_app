package main

import (
	"log"
	"net/http"
	"time"

	"github.com/homie-dev/gotwiml/twiml"
	"github.com/homie-dev/gotwiml/twiml/attr"
	"github.com/sfreiberg/gotwilio"
)

const (
	// Twilio管理画面より取得
	accountSID        = "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	apiKeySID         = "SKXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	apiKeySecret      = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	applicationSID    = "APXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	twilioPhoneNumber = "+1XXXYYYZZZZ"

	// ログイン名
	identity = "planet-meron"
)

// 着信用のTwiML生成
func genIncomingTwiml() ([]byte, error) {
	resp := twiml.NewVoiceResponse().
		AppendDial(twiml.NewDial().Client(identity))
	xml, err := resp.ToXML()
	if err != nil {
		return nil, err
	}
	return []byte(xml), nil
}

// 発信用のTwiML生成
func genOutgoingTwiml(phoneNumber string) ([]byte, error) {
	resp := twiml.NewVoiceResponse().
		AppendDial(twiml.NewDial(
			// 発信元電話番号
			attr.CallerID(twilioPhoneNumber),
		).
			Number(phoneNumber),
		)
	xml, err := resp.ToXML()
	if err != nil {
		return nil, err
	}
	return []byte(xml), nil
}

// 初期化用のアクセストークンの生成
func genCallToken() (string, error) {
	// Twilioクライアントの初期化
	twilio := gotwilio.Twilio{
		AccountSid:   accountSID,
		APIKeySid:    apiKeySID,
		APIKeySecret: apiKeySecret,
	}

	// アクセストークンの初期化と権限設定
	tk := twilio.NewAccessToken()
	tk.AddGrant(gotwilio.VoiceGrant{
		Incoming: gotwilio.VoiceGrantIncoming{Allow: true}, // 着信用に追加
		Outgoing: gotwilio.VoiceGrantOutgoing{
			ApplicationSID: applicationSID,
		},
	})

	// デバイスの識別名(着信などに利用)
	tk.Identity = identity

	// アクセストークンの有効期限を1時間に設定
	tk.ExpiresAt = time.Now().Add(1 * time.Hour)
	return tk.ToJWT()
}

func main() {
	// public配下の静的ファイルを公開
	http.Handle("/", http.FileServer(http.Dir("public")))

	// アクセストークンの生成
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		token, err := genCallToken()
		if err != nil {
			panic(err)
		}
		if _, err := w.Write([]byte(token)); err != nil {
			panic(err)
		}
	})

	// 発信用TwiMLの生成
	http.HandleFunc("/outgoing", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			panic(err)
		}
		// 発信先の電話番号の取得
		to := r.FormValue("To")

		// 発信先へ電話をかけるTwiMLの生成
		resp, err := genOutgoingTwiml(to)
		if err != nil {
			panic(err)
		}

		if _, err := w.Write(resp); err != nil {
			panic(err)
		}
	})

	// 着信用TwiMLの生成のエンドポイント
	http.HandleFunc("/incoming", func(w http.ResponseWriter, r *http.Request) {
		resp, err := genIncomingTwiml()
		if err != nil {
			panic(err)
		}
		if _, err := w.Write(resp); err != nil {
			panic(err)
		}
	})

	// サーバーの起動
	if err := http.ListenAndServe(":8686", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
