package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/atotto/clipboard"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"gopkg.in/toast.v1"
)

// ==========================================
// 1. 認証関連のヘルパー関数
// ==========================================

// getClient は設定情報を元にHTTPクライアントを生成します（初回はWeb認証）
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// getTokenFromWeb はブラウザでの認証を促し、トークンを取得します
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("以下のURLをブラウザで開き、ログインして認証コードを取得してください:\n\n%v\n\n", authURL)
	fmt.Print("取得した認証コードをここに貼り付けてEnter: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("コードの読み取りに失敗しました: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("トークンの取得に失敗しました: %v", err)
	}
	return tok
}

// tokenFromFile は保存された token.json からトークンを読み込みます
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken は取得したトークンを token.json に保存します
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("トークンを保存しています: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("トークンの保存に失敗しました: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// ==========================================
// 2. メインの処理とループ
// ==========================================
func main() {
	ctx := context.Background()

	// credentials.json の読み込み
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("credentials.json の読み込みエラー: %v", err)
	}

	// Gmail APIの権限（読み取りのみでOKなため ReadOnlyScope を指定）
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("設定の解析エラー: %v", err)
	}
	client := getClient(config)

	// Gmailサービスをビルド
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Gmailクライアント作成エラー: %v", err)
	}

	fmt.Println("認証成功！監視をスタートしました...")

	// 抽出用の正規表現をコンパイル
	re := regexp.MustCompile(`認証コードは\s*(.+?)\s*です`)
	// 処理済みメッセージIDを保存するマップ
	processedIDs := make(map[string]bool)
	// 10秒ごとに実行するティッカー（タイマー）を作成
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// 無限ループで定期実行
	for range ticker.C {
		checkAuthCode(srv, re, processedIDs)
	}
}

// checkAuthCode はメールの検索とクリップボードへのコピーを行います
func checkAuthCode(srv *gmail.Service, re *regexp.Regexp, processedIDs map[string]bool) {
	user := "me"

	// 未読で、件名に「認証コードは」を含むメールを検索
	r, err := srv.Users.Messages.List(user).Q(`is:unread subject:"認証コードは"`).Do()
	if err != nil {
		log.Printf("検索エラー: %v", err)
		return
	}

	for _, msg := range r.Messages {
		if processedIDs[msg.Id] {
			// 既に処理済みのmsgはスキップ
			continue
		}

		// ヘッダー情報（件名）のみを軽量に取得
		m, err := srv.Users.Messages.Get(user, msg.Id).Format("metadata").MetadataHeaders("Subject").Do()
		if err != nil {
			continue
		}

		var subject string
		for _, header := range m.Payload.Headers {
			if header.Name == "Subject" {
				subject = header.Value
				break
			}
		}

		// 正規表現でコードを抜き出す
		matches := re.FindStringSubmatch(subject)
		if len(matches) > 1 {
			authCode := matches[1]

			// クリップボードにコピー
			err := clipboard.WriteAll(authCode)
			if err != nil {
				log.Printf("コピー失敗: %v", err)
			} else {
				fmt.Printf("★ コピー成功: %s (件名: %s)\n", authCode, subject)

				// 処理済みとして記憶
				processedIDs[msg.Id] = true

				// ==========================================
				// ★追加: トースト通知を送信
				// ==========================================
				notification := toast.Notification{
					AppID:   "Copy Auth Code",                  // 通知の送信元として表示される名前
					Title:   "Code Copyed",                     // 通知のタイトル
					Message: fmt.Sprintf("Code: %s", authCode), // 通知の本文
					Audio:   toast.Default,                     // ※コメントを外すと通知音を変えられます
				}

				err = notification.Push()
				if err != nil {
					log.Printf("通知の送信に失敗しました: %v", err)
				}
			}
		}
	}
}
