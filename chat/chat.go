package chat

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/assistent-ai/client/gpt"
	"github.com/assistent-ai/client/model"

	"github.com/assistent-ai/client/db"
	"github.com/google/uuid"
)

func ShowMessages(messages []model.Message) {
	for _, message := range messages {
		formattedTimestamp := message.Timestamp.Format(model.TimestampFormattingTemplate)
		fmt.Printf("[%s] %s: %s\n", formattedTimestamp, message.Role, message.Content)
	}
}

func StartChat(dialogId string, ctx *model.AppContext) error {
	if dialogId == "" {
		dialogId = model.DefaultDialogId
	}

	// Create a new scanner to read messages from the user
	scanner := bufio.NewScanner(os.Stdin)
	messages, err := db.GetMessagesByDialogID(dialogId)
	if err != nil {
		return err
	}
	ShowMessages(messages)

	for {
		// Print a prompt to the user
		fmt.Print("You: ")

		// Read a line of text from the user
		if !scanner.Scan() {
			// If there was an error reading input, break out of the loop
			break
		}
		msgUUID, err := uuid.NewRandom()
		msgId := msgUUID.String()

		newMessage := model.Message{
			ID:        msgId,
			DialogId:  dialogId,
			Timestamp: time.Now(),
			Role:      model.UserRoleName,
			Content:   scanner.Text(),
		}
		messages = append(messages, newMessage)
		over, err := gpt.IsDialogOver(messages, ctx)
		if err != nil && over {
			break
		}
		messages := gpt.TrimMessages(messages, 10000)
		messages, err = gpt.Message(messages, dialogId, ctx)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}
		if _, err = db.StoreMessage(newMessage); err != nil {
			return err
		}
		lastMessage := messages[len(messages)-1]
		if _, err = db.StoreMessage(lastMessage); err != nil {
			return err
		}

		// Print the last message
		fmt.Printf("%s: %s\n", lastMessage.Role, lastMessage.Content)
	}

	// If we've reached the end of input, print a goodbye message
	fmt.Println("Goodbye!")
	return nil
}
