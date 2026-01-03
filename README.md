# Wakaba

Discord Link Summarizer Bot for AWS Lambda.

## 機能
- `/summarize date:MMDD`
指定された日付 (MMDD) または (YYYYMMDD) に投稿された URL を抽出してまとめます

## デプロイ手順
### 前提条件
- Go 1.2x
- Make
- Zip

### 1. ビルド～zip化
```bash
make
```

### 2. AWS Lambda のセットアップ
1. **関数の作成**:
   - ランタイム: `provided.al2023` (Amazon Linux 2023)
   - アーキテクチャ: `x86_64`
2. **コードのアップロード**:
   - 生成された `function.zip` をアップロード
3. **環境変数の設定**:
   - `DISCORD_PUBLIC_KEY`: Discord Bot の Public Key
   - `DISCORD_BOT_TOKEN`: Discord Bot Token
   - `DISCORD_APP_ID`: Application ID (コマンド登録時に使用)
4. **IAM ロールの設定**:
   - Lambda が自分自身を再帰呼び出しするために、実行ロールに `lambda:InvokeFunction` 権限を追加する必要があります。
   - インラインポリシー例:
     ```json
     {
       "Version": "2012-10-17",
       "Statement": [
         {
           "Effect": "Allow",
           "Action": "lambda:InvokeFunction",
           "Resource": "*"
         }
       ]
     }
     ```

### 3. API Gateway の設定
1. **API の作成**:
   - HTTP API (または REST API) を作成します。
2. **統合**:
   - Lambda 関数と統合します。
3. **Endpoint URL の設定**:
   - 生成された API の URL を Discord Developer Portal の "Interactions Endpoint URL" に設定します。

## コマンドの登録

Bot をサーバーに追加しただけではスラッシュコマンドは使用できません。以下の手順でコマンドを登録してください。

### 手順
環境変数を設定した上で、`make register` を実行します。

```bash
export DISCORD_BOT_TOKEN="your_bot_token"
export DISCORD_APP_ID="your_application_id"

# 特定のサーバーに登録する場合（即時反映・推奨）
make register GUILD_ID="your_guild_id"

# または、全サーバー（グローバル）に登録する場合（反映に最大1時間）
# make register
```

## 開発

### ローカルビルド
```bash
make build
```

### クリーンアップ
```bash
make clean
```
