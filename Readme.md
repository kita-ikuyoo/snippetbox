# Snippetbox - A web application for saving and sharing text snippets built with Go

## How to use
```bash
docker-compose up
```
コンテナが建てた後、localhost:443にアクセスすることで確認できます。

## About


# Snippetbox — Go製 Webアプリケーション

Alex Edwards著「Let's Go」を完走して構築した、スニペット共有WebアプリのGoによる実装です。

🔗 [ライブデモ](https://your-app-url.com) ← あれば

---

## 技術スタック

| カテゴリ   | 技術                   |
|--------|----------------------|
| 言語     | Go 1.22              |
| Web    | net/http（標準ライブラリ）    |
| DB     | MySQL + database/sql |
| テンプレート | html/template        |
| 認証     | セッション管理 + bcrypt     |
| テスト    | testing（標準ライブラリ）     |
| 実行環境   | docker               |


---

## 主な機能

- スニペットの作成・閲覧
- ユーザー登録・ログイン（bcryptパスワードハッシュ）
- SSL/TLSウェブサーバー
- MySQLデータベースを用いたデータの永続化
- セッション管理（ステートフルHTTP）
- CSRFトークンによるセキュリティ対策
- ミドルウェアチェーンによるリクエスト処理
- テーブル駆動テスト・統合テスト

---

## こだわった点

- **依存性注入**: `application` 構造体にDBやロガーをまとめ、グローバル変数を排除
- **ミドルウェア設計**: セキュリティヘッダー・ロギング・パニックリカバリーを分離して構成
- **安全性設計**: SSL/TLSとCSRFトークンによる安全性対策
- **エラーハンドリング**: 集中管理により各ハンドラをシンプルに保った
- **テスト**: `httptest` を使ったエンドツーエンドテストを実装

---

## セットアップ
```bash
git clone https://github.com/[yourname]/snippetbox
cd snippetbox

# DBの準備
mysql -u root < db/setup.sql

# 環境変数
export DSN="web:pass@/snippetbox?parseTime=true"

# 起動
go run ./cmd/web
# → http://localhost:4000
```

---

## 学んだこと

本書を通じて、GoのHTTPサーバーの仕組みをフレームワーク無しで理解しました。
特にミドルウェアの実装原理、テンプレートキャッシュ、セッション管理の設計を
自分の手で組み上げたことで、Goらしい設計思想を体感できました。