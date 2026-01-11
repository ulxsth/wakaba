package discord

import (
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yotu/wakaba/internal/util"
)

// 該当する日付のメッセージを取得する結果を示す構造体
type FetchResult struct {
	CapturedLinks []string
	MessageCount  int
}

// 該当する日付のメッセージを取得する
func FetchLinks(s *discordgo.Session, channelID string, start, end time.Time, botID string) (*FetchResult, error) {
	var messages []*discordgo.Message
	var lastID string

	for {
		// 一度に最大 100 件のメッセージを取得
		batch, err := s.ChannelMessages(channelID, 100, lastID, "", "")
		if err != nil {
			return nil, err
		}

		if len(batch) == 0 {
			break
		}

		lastID = batch[len(batch)-1].ID

		// 取得したメッセージを処理
		for _, m := range batch {
			// discord Snowflake ID からメッセージの作成時間を逆算
			ts, err := discordgo.SnowflakeTimestamp(m.ID)
			if err != nil {
				ts = m.Timestamp
			}

			// 終了日時よりも新しい場合は、スキップ
			if ts.After(end) {
				continue
			}

			// 開始日時よりも古い場合は、取得を終了
			if ts.Before(start) {
				goto DONE
			}

			// ボットのメッセージは無視
			if m.Author.ID == botID {
				continue
			}

			messages = append(messages, m)
		}
	}

// すべて取得し終わったら、ソートしてリンクを抽出
DONE:
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})

	var links []string
	for _, m := range messages {
		extracted := util.ExtractURLs(m.Content)
		links = append(links, extracted...)
	}

	return &FetchResult{
		CapturedLinks: links,
		MessageCount:  len(messages),
	}, nil
}
