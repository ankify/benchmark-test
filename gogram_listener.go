package main

import (
	"fmt"
	"os"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func main() {
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:    os.Getenv("API_ID"),
		AppHash: os.Getenv("API_HASH"),
		BotToken: os.Getenv("BOT_TOKEN"),
	})

	if err != nil {
		panic(err)
	}

	client.OnNewMessage(func(m *telegram.NewMessage) error {
		if m.Message.Document == nil {
			return nil
		}

		sizeMB := float64(m.Message.Document.Size) / (1024 * 1024)
		if sizeMB < 100 {
			return nil
		}

		client.SendMessage(m.ChatID, "📥 Gogram: Downloading...")

		file := "gogram_" + m.Message.Document.FileName

		// ⬇️ Download
		start := time.Now()
		err := client.DownloadMedia(m.Message, file)
		if err != nil {
			return err
		}
		downloadTime := time.Since(start)

		client.SendMessage(m.ChatID, "📤 Gogram: Uploading...")

		// ⬆️ Upload
		start = time.Now()
		_, err = client.SendDocument(m.ChatID, file, nil)
		if err != nil {
			return err
		}
		uploadTime := time.Since(start)

		client.SendMessage(
			m.ChatID,
			fmt.Sprintf(
				"✅ **Gogram Result**\n📦 Size: `%.2f MB`\n⬇️ Download: `%v`\n⬆️ Upload: `%v`",
				sizeMB,
				downloadTime,
				uploadTime,
			),
		)

		return nil
	})

	client.Start()
	select {}
}
