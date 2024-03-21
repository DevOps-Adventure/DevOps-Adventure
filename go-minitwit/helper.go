package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

// Helper functions
func checkPasswordHash(userEnteredPwd string, dbpwd string) bool {
	bytes := md5.Sum([]byte(userEnteredPwd))
	str := hex.EncodeToString(bytes[:])
	return str == dbpwd
}

func gravatarURL(email string, size int) string {
	if size <= 0 {
		size = 80 // Default size
	}

	email = strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(email))
	return fmt.Sprintf("http://www.gravatar.com/avatar/%x?d=identicon&s=%d", hash, size)
}

// package messages to be displayed on the UI
func formatMessages(messages []MessageUser) []MessageUI {
	var formattedMessages []MessageUI
	for _, m := range messages {
		var msg MessageUI

		msg.MessageID = int(m.MessageID)
		msg.AuthorID = int(m.AuthorID)
		msg.User.UserID = int(m.UserID)
		msg.Text = string(m.Text)
		msg.Username = string(m.Username)
		msg.Email = string(m.Email)

		pubDateTime := time.Unix(int64(m.PubDate), 0)
		msg.PubDate = pubDateTime.Format("02/01/2006 15:04:05")

		link := "/" + msg.Username
		msg.Profile_link = strings.ReplaceAll(link, " ", "%20")

		gravatarURL := gravatarURL(msg.Email, 48)
		msg.Gravatar = gravatarURL

		formattedMessages = append(formattedMessages, msg)
	}

	return formattedMessages
}

// package messages to be sent back to the API
func filterMessages(messages []MessageUser) []FilteredMsg {
	var filteredMessages []FilteredMsg
	for _, m := range messages {
		var msg FilteredMsg
		// content
		msg.Content = string(m.Text)

		// publication date
		msg.PubDate = int64(m.PubDate)

		// user
		msg.User = string(m.Username)

		filteredMessages = append(filteredMessages, msg)
	}
	return filteredMessages
}

func logMessage(message string) {
	// Specify the file path
	filePath := "./tmp/logging/logger.txt"

	// Open or create the file for writing
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	data := []byte(message)

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
