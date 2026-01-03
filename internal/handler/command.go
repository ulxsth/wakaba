package handler

import (
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
	DateArg          string `json:"date_arg"`
	WithTitle        bool   `json:"with_title"`
}

// リクエストに基づき、コマンドの本処理を実行する
func ProcessCommand(req *SummaryRequest, botToken string) error {
	// session 作成
	s, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}

	// 日付引数をパース
	now := time.Now()
	start, end, err := util.ParseDateInput(req.DateArg, now)
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

	// Format Output
	// New format:
	// count: n
	// ```
	// Title
	// URL
	// ...
	// ```

	countHeader := fmt.Sprintf("count: %d\n", len(result.CapturedLinks))
	var sb strings.Builder
	sb.WriteString(countHeader)
	sb.WriteString("```\n")

	if req.WithTitle {
		// Parallel fetching
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
					// Fallback to empty title or domain?
					// Prompt asked for:
					// title1
					// url1
					// If title missing, maybe just empty line or skip?
					// Let's use empty string or "(no title)"
					t = ""
				}
				ch <- titleResult{index: i, title: t, url: u}
			}(i, u)
		}

		// Collect results
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

	} else {
		// Just URLs
		sb.WriteString(strings.Join(result.CapturedLinks, "\n"))
		sb.WriteString("\n")
	}

	sb.WriteString("```")
	content := sb.String()

	// Discord message limit is 2000 chars. Simple split if needed, or truncation.
	if len(content) > 2000 {
		// Truncate and ensure we close the code block
		// 1900 gives us enough buffer for the suffix
		content = content[:1900] + "\n...(省略されました)\n```"
	}

	return sendFollowup(s, req, content)
}

func sendError(s *discordgo.Session, req *SummaryRequest, msg string) error {
	return sendFollowup(s, req, "エラー: "+msg)
}

func sendFollowup(s *discordgo.Session, req *SummaryRequest, content string) error {
	// WebhookEditWithToken is used for "Deferred" interaction responses to "EditOriginal".
	// Or WebhookExecute for followup?
	// If we deferred, we can use InteractioResponseEdit (which updates the "Loading..." message)
	// OR FollowupMessageCreate.

	// Use WebhookMessageEdit to update the original deferred message.
	// The webhook ID is the Application ID, and the token is the Interaction Token.
	// Message ID "@original" targets the initial response.
	_, err := s.WebhookMessageEdit(req.ApplicationID, req.InteractionToken, "@original", &discordgo.WebhookEdit{
		Content: &content,
	})

	if err != nil {
		log.Printf("Failed to send followup: %v", err)
		return err
	}
	return nil
}
