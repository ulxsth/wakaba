package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/yotu/wakaba/internal/repository"
)

const (
	PageSize = 10
)

func ProcessTodoList(s *discordgo.Session, req *WorkerRequest) error {
	repo, err := repository.NewTodoRepository(context.Background())
	if err != nil {
		return sendError(s, req, fmt.Sprintf("Repository init failed: %v", err))
	}

	var args TodoListArgs
	if argBytes, err := json.Marshal(req.CommandArgs); err == nil {
		json.Unmarshal(argBytes, &args)
	}

	switch args.SubCommand {
	case "create":
		return handleCreateList(s, req, repo)
	case "add":
		return handleAddItem(s, req, repo, args.Content)
	default:
		return sendError(s, req, "Unknown subcommand")
	}
}

func handleCreateList(s *discordgo.Session, req *WorkerRequest, repo *repository.TodoRepository) error {
	list, err := repo.GetTodoList(context.Background(), req.ChannelID)
	if err != nil {
		return sendError(s, req, fmt.Sprintf("Failed to get list: %v", err))
	}

	if list.MessageID != "" {
		// If a list already exists, maybe we should delete the old message?
		// For now, just post a new one and update the ID.
	}

	// Create initial message
	embed, components := renderTodoList(list, 0)

	// If invoking via slash command, use InteractionResponse
	// But we are in a worker, so we edit the deferred response or send a new message.
	// Since Slash Command "create" expects a response, we should reply to it.

	// However, usually we want the TODO list to be a persistent message.
	// Slash Command response is ephemeral or tied to the interaction.
	// We want to send a NEW message to the channel and track it.

	embeds := []*discordgo.MessageEmbed{embed}
	msg, err := s.ChannelMessageSendComplex(req.ChannelID, &discordgo.MessageSend{
		Embeds:     embeds,
		Components: components,
	})
	if err != nil {
		return sendError(s, req, fmt.Sprintf("Failed to send list message: %v", err))
	}

	list.MessageID = msg.ID
	if err := repo.SaveTodoList(context.Background(), list); err != nil {
		return sendError(s, req, fmt.Sprintf("Failed to save list: %v", err))
	}

	return sendFollowup(s, req, "TODOリストを作成しました。")
}

func handleAddItem(s *discordgo.Session, req *WorkerRequest, repo *repository.TodoRepository, content string) error {
	list, err := repo.GetTodoList(context.Background(), req.ChannelID)
	if err != nil {
		return sendError(s, req, fmt.Sprintf("Failed to get list: %v", err))
	}

	if list.MessageID == "" {
		return sendError(s, req, "TODOリストがありません。先に `/list create` を実行してください。")
	}

	newItem := repository.TodoItem{
		ID:      fmt.Sprintf("%d", len(list.Items)+1), // Simple ID generation
		Content: content,
		Status:  "open",
	}
	list.Items = append(list.Items, newItem)

	if err := repo.SaveTodoList(context.Background(), list); err != nil {
		return sendError(s, req, fmt.Sprintf("Failed to save list: %v", err))
	}

	// Update the pinned message
	if err := updateListMessage(s, req.ChannelID, list, 0); err != nil {
		return sendError(s, req, fmt.Sprintf("Failed to update message: %v", err))
	}

	return sendFollowup(s, req, fmt.Sprintf("タスクを追加しました: %s", content))
}

func ProcessTodoComponent(s *discordgo.Session, req *WorkerRequest) error {
	repo, err := repository.NewTodoRepository(context.Background())
	if err != nil {
		// Component interaction failures are silent to the user usually unless we update the message
		log.Printf("Repository init failed: %v", err)
		return nil
	}

	list, err := repo.GetTodoList(context.Background(), req.ChannelID)
	if err != nil {
		log.Printf("Failed to get list: %v", err)
		return nil
	}

	// CustomID format: "todo:action:page:extra"
	parts := strings.Split(req.CustomID, ":")
	if len(parts) < 3 || parts[0] != "todo" {
		return nil
	}

	action := parts[1]
	var page int
	fmt.Sscanf(parts[2], "%d", &page)

	switch action {
	case "prev":
		page--
		if page < 0 {
			page = 0
		}
	case "next":
		page++
	case "complete":
		if len(parts) >= 4 {
			idxStr := parts[3]
			var idx int
			fmt.Sscanf(idxStr, "%d", &idx)

			// Find item with ID = idxStr (Assuming ID is string representation of index initially, but keeping it simple)
			// Wait, in handleAddItem, ID is "%d".
			// We should iterate.
			for i, item := range list.Items {
				if item.ID == idxStr {
					if list.Items[i].Status == "done" {
						list.Items[i].Status = "open"
					} else {
						list.Items[i].Status = "done"
					}
					repo.SaveTodoList(context.Background(), list)
					break
				}
			}
		}
	}

	return updateListMessage(s, req.ChannelID, list, page)
}

func updateListMessage(s *discordgo.Session, channelID string, list *repository.TodoList, page int) error {
	embed, components := renderTodoList(list, page)

	embeds := []*discordgo.MessageEmbed{embed}
	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    channelID,
		ID:         list.MessageID,
		Embeds:     &embeds,
		Components: &components,
	})
	return err
}

func renderTodoList(list *repository.TodoList, page int) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	totalItems := len(list.Items)
	totalPages := (totalItems + PageSize - 1) / PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * PageSize
	end := start + PageSize
	if end > totalItems {
		end = totalItems
	}

	var description strings.Builder
	var rowButtons []discordgo.MessageComponent

	for i := start; i < end; i++ {
		item := list.Items[i]
		statusIcon := "⬜"
		if item.Status == "done" {
			statusIcon = "✅"
		}
		description.WriteString(fmt.Sprintf("%s %s\n", statusIcon, item.Content))

		// Number button
		style := discordgo.SecondaryButton
		if item.Status == "done" {
			style = discordgo.SuccessButton
		}

		// Button ID: todo:complete:page:itemID
		btn := discordgo.Button{
			Label:    fmt.Sprintf("%d", i+1), // Display 1-based index roughly? Or just number them 1-10 on the page? Issue says 1-10.
			CustomID: fmt.Sprintf("todo:complete:%d:%s", page, item.ID),
			Style:    style,
		}
		// Issue says: 1-10のリアクションは、該当するページのタスクの番号に対応する
		// So buttons should be numbered 1, 2, 3... relative to the page OR relative to the list?
		// "1-10のリアクション" implies 1-10 digits.
		// Let's use 1-10 labels map to item.ID.
		btn.Label = fmt.Sprintf("%d", (i%PageSize)+1)

		rowButtons = append(rowButtons, btn)
	}

	if description.Len() == 0 {
		description.WriteString("（タスクはありません）")
	}

	embed := &discordgo.MessageEmbed{
		Title:       "TODO リスト",
		Description: description.String(),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d/%d (%d items)", page+1, totalPages, totalItems),
		},
		Color: 0x00ff00,
	}

	// Pagination buttons
	// navRow := discordgo.ActionsRow{ ... } // Unused logic removed

	// Better nav row: [Prev] [1 2 3 4 5] [Next] ?
	// Max components per row is 5.
	// We can have up to 10 item buttons. That takes 2 rows of 5.
	// And 1 row for navigation.

	components := []discordgo.MessageComponent{}

	// Item buttons rows
	// Slice rowButtons into chunks of 5
	for i := 0; i < len(rowButtons); i += 5 {
		end := i + 5
		if end > len(rowButtons) {
			end = len(rowButtons)
		}
		components = append(components, discordgo.ActionsRow{
			Components: rowButtons[i:end],
		})
	}

	// Navigation Row
	navComponents := []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "⬅️ 前へ",
			CustomID: fmt.Sprintf("todo:prev:%d", page),
			Style:    discordgo.PrimaryButton,
			Disabled: page == 0,
		},
		discordgo.Button{
			Label:    "次へ ➡️",
			CustomID: fmt.Sprintf("todo:next:%d", page),
			Style:    discordgo.PrimaryButton,
			Disabled: page >= totalPages-1,
		},
	}
	components = append(components, discordgo.ActionsRow{
		Components: navComponents,
	})

	return embed, components
}
