package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/assistent-ai/client/model"
	"github.com/b0noi/go-utils/v2/gcp"
	"github.com/google/uuid"
)

func IsDialogOver(messages []model.Message) (bool, error) {
	apiKey, err := gcp.AccessSecretVersion("projects/16255416068/secrets/gpt3-secret/versions/1")
	if err != nil {
		return false, err
	}

	url := "https://api.openai.com/v1/chat/completions"

	// Create a new chat.Message with the GPT-4 response
	newMessage := model.Message{
		ID:        "", // You can assign a new ID here
		DialogId:  "", // You can assign a new DialogId here
		Timestamp: time.Now(),
		Role:      model.SystemRoleName,
		Content:   "Based on the last response from the user, is this dialog over? Please respond with true/false only",
	}

	messages = append(messages, newMessage)

	requestBody, err := prepareGPT4RequestBody(messages)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	// var response map[string]interface{}
	var response GptChatCompletionMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, err
	}
	result, err := strconv.ParseBool(response.Choices[0].Message.Content)
	if err != nil {
		return false, err
	}
	return result, nil
}

func Message(messages []model.Message, dialogId string) ([]model.Message, error) {
	apiKey, err := gcp.AccessSecretVersion("projects/16255416068/secrets/gpt3-secret/versions/1")
	if err != nil {
		return nil, err
	}

	url := "https://api.openai.com/v1/chat/completions"

	requestBody, err := prepareGPT4RequestBody(messages)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// var response map[string]interface{}
	var response GptChatCompletionMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return addGPT4Response(response, messages, dialogId)
}

func addGPT4Response(response GptChatCompletionMessage, messages []model.Message, dialogId string) ([]model.Message, error) {
	// Assume we're only getting 1 response, so we use the first choice
	gpt4Text := response.Choices[0].Message.Content
	uuidMsg, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	idMsg := uuidMsg.String()

	// Create a new chat.Message with the GPT-4 response
	newMessage := model.Message{
		ID:        idMsg,    // You can assign a new ID here
		DialogId:  dialogId, // You can assign a new DialogId here
		Timestamp: time.Now(),
		Role:      model.AssistentRoleNeam,
		Content:   gpt4Text,
	}

	messages = append(messages, newMessage)

	// Append the new message to the input messages slice
	return messages, nil
}

func prepareGPT4RequestBody(messages []model.Message) ([]byte, error) {
	// Create a new slice to hold message maps
	gptMessages := make([]map[string]string, len(messages))

	// Iterate through the input messages
	for i, message := range messages {
		// Convert the timestamp to a human-readable format
		formattedTimestamp := message.Timestamp.Format("2006-01-02 15:04:05")

		// Combine the content with the timestamp
		combinedContent := fmt.Sprintf("%s: %s", formattedTimestamp, message.Content)

		// Add the message to the gptMessages slice
		gptMessages[i] = map[string]string{
			"role":    message.Role,
			"content": combinedContent,
		}
	}

	// Marshal the request body for GPT-4
	requestBody, err := json.Marshal(map[string]interface{}{
		"messages":   gptMessages,
		"max_tokens": 2000,
		"n":          1,
		"model":      "gpt-4",
	})

	if err != nil {
		return nil, err
	}

	return requestBody, nil
}
