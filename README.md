# AutoAuth

Gmailに届く「認証コード」を自動的に読み取り、クリップボードへコピー＆Windowsトースト通知でお知らせする常駐型ツールです。

## ✨ 特徴 (Features)

* **完全自動化**: 起動しておくだけで、新しく届いた認証コードを自動でクリップボードにコピーします。
* **Windowsネイティブ通知**: コピー成功時にWindowsのトースト通知でお知らせします。
* **セキュアな設計**: Gmail APIの「読み取り専用（`Read-only`）」スコープを使用。メールの削除や変更は一切行えない安全な設計です。
* **バックグラウンド動作**: コンソール画面（黒い画面）を出さずに裏側でひっそりと常駐します。

## 🛠️ 技術スタック

* **言語**: Go (Golang) v1.26.2
* **API**: Gmail API (`google.golang.org/api/gmail/v1`)
* **OS**: Windows 11 (開発環境: WSL2)

## 🚀 使い方 (Usage)

### 1. 実行ファイルのダウンロード
[Releases](../../releases) ページから、最新の `autoauth_bg.exe`（常駐版）または `autoauth.exe`（コンソール版）をダウンロードします。

### 2. credentials.json の準備
本アプリを利用するには、ご自身のGoogleアカウントでAPIを有効化し、認証情報を取得する必要があります。

1. [Google Cloud Console](https://console.cloud.google.com/) にアクセスします。
2. 新しいプロジェクトを作成し、「Gmail API」を有効化します。
3. 「OAuth 同意画面」を設定します（テストユーザーにご自身のGmailアドレスを追加）。
4. 「認証情報」から「OAuth クライアント ID（デスクトップ アプリ）」を作成し、JSONをダウンロードします。
5. ファイル名を `credentials.json` に変更し、ダウンロードした `.exe` ファイルと同じフォルダに配置します。

### 3. 初回起動と認証
コマンドプロンプト等からアプリを起動します（初回はコンソール版の `autoauth.exe` を推奨します）。

```bash
autoauth.exe
```
コンソールに認証用のURLが表示されます。ブラウザで開き、Googleアカウントへのアクセス（読み取り専用）を許可してください。表示されたコードをコンソールに貼り付けると `token.json` が生成され、監視がスタートします。

### 4. 常駐化（バックグラウンド実行）
`token.json` が生成された後は、`autoauth_bg.exe` をダブルクリックするだけで、画面を出さずにバックグラウンドで監視を続けます。
終了したい場合は、Windowsの「タスクマネージャー」から `autoauth_bg.exe` を終了させてください。

## ⚙️ 開発者向け (Build from source)

ご自身でビルドを行う場合は、以下のコマンドを実行してください。

```bash
# パッケージのインストール
go mod download

# Windows用コンソール版のビルド
GOOS=windows GOARCH=amd64 go build -o autoauth.exe main.go

# Windows用バックグラウンド版（画面なし）のビルド
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o autoauth_bg.exe main.go
```

## ⚠️ セキュリティに関する注意事項

* `credentials.json` および生成される `token.json` には、あなたのアカウントにアクセスするための重要な情報が含まれています。
* **絶対にこれらのファイルをGitHubなどの公開リポジトリにコミット（Push）しないでください。**
* リポジトリには `.gitignore` が設定されており、デフォルトでこれらのファイルは除外されます。

## 📄 ライセンス
MIT License
