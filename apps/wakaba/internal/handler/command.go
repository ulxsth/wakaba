package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yotu/wakaba/internal/discord"
	"github.com/yotu/wakaba/internal/util"
)

// 本処理時のリクエストを示すデータ形式
type SummaryRequest struct {
	InteractionID    string `json:"interaction_id"`
	InteractionToken string `json:"interaction_token"`
	ChannelID        string `json:"channel_id"`
	ApplicationID    string `json:"application_id"`

	// 以下ユーザ入力のコマンド引数
	DateArg   string `json:"date_arg"`
	WithTitle bool   `json:"with_title"`
}

// リクエストに基づき、コマンドの本処理を実行する
func ProcessSummarize(s *discordgo.Session, req *WorkerRequest) error {
	// Args map to struct
	var args SummarizeArgs
	// Marshaling round trip is a lazy way to map map[string]any to struct, but works
	if argBytes, err := json.Marshal(req.CommandArgs); err == nil {
		json.Unmarshal(argBytes, &args)
	}

	// 日付引数をパース
	now := time.Now()
	start, end, err := util.ParseDateInput(args.DateArg, now)
	if err != nil {
		return sendError(s, req, fmt.Sprintf("日付の形式が正しくありません: %v", err))
	}

	// 該当する日付のメッセージを取得
	result, err := discord.FetchLinks(s, req.ChannelID, start, end, req.ApplicationID)
	if err != nil {
		return sendError(s, req, fmt.Sprintf("メッセージの取得に失敗しました: %v", err))
	}

	if len(result.CapturedLinks) == 0 {
		return sendFollowup(s, req, fmt.Sprintf("%s のリンクは見つかりませんでした。(検索数: %d件)", start.Format("2006/01/02"), result.MessageCount))
	}

	countHeader := fmt.Sprintf("count: %d\n", len(result.CapturedLinks))
	var sb strings.Builder
	sb.WriteString(countHeader)
	sb.WriteString("```\n")

	// with_title = True の場合は url のタイトルを取得して表示する
	if args.WithTitle {
		type titleResult struct {
			index int
			title string
			url   string
		}
		ch := make(chan titleResult, len(result.CapturedLinks))

		for i, u := range result.CapturedLinks {
			go func(i int, u string) {
				t, err := util.FetchPageTitle(u)
				if err != nil {
					t = "(no title)"
				}
				ch <- titleResult{index: i, title: t, url: u}
			}(i, u)
		}

		results := make([]titleResult, len(result.CapturedLinks))
		for i := 0; i < len(result.CapturedLinks); i++ {
			r := <-ch
			results[r.index] = r
		}

		for _, r := range results {
			if r.title != "" {
				sb.WriteString(r.title + "\n")
			}
			sb.WriteString(r.url + "\n\n")
		}

		// そうでない場合はurlのみ表示
	} else {
		sb.WriteString(strings.Join(result.CapturedLinks, "\n"))
		sb.WriteString("\n")
	}

	sb.WriteString("```")
	content := sb.String()

	// 2000文字を超えた場合は省略
	if len(content) > 2000 {
		content = content[:1900] + "\n...(略)\n```"
	}

	return sendFollowup(s, req, content)
}

func sendError(s *discordgo.Session, req *WorkerRequest, msg string) error {
	return sendFollowup(s, req, "エラー: "+msg)
}

// 処理結果を表示する（元のメッセージを更新する形で送信する）
func sendFollowup(s *discordgo.Session, req *WorkerRequest, content string) error {
	// WebhookMessageEdit は指定したメッセージを更新する
	// messageId = @original は slash command に対する最初のレスポンス（ping-pong 時に表示される「考え中...」）を指す
	_, err := s.WebhookMessageEdit(req.ApplicationID, req.InteractionToken, "@original", &discordgo.WebhookEdit{
		Content: &content,
	})

	if err != nil {
		log.Printf("Failed to send followup: %v", err)
		return err
	}
	return nil
}
