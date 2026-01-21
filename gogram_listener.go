package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func mustInt32(s string) int32 {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return int32(i)
}

func main() {
	apiID := mustInt32(os.Getenv("API_ID"))
	apiHash := os.Getenv("API_HASH")
	botToken := os.Getenv("BOT_TOKEN")

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:   apiID,
		AppHash: apiHash,
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// connect then login bot
	if err := client.Connect(); err != nil {
		log.Fatalf("connect failed: %v", err)
	}
	if err := client.LoginBot(botToken); err != nil {
		log.Fatalf("bot login failed: %v", err)
	}

	// Message handler
	client.On(telegram.OnMessage, func(m *telegram.NewMessage) error {
		// ignore non-media
		if !m.IsMedia() {
			return nil
		}

		// only handle documents
		doc := m.Document()
		if doc == nil {
			return nil
		}

		// Size field exists on DocumentObj
		sizeMB := float64(doc.Size) / (1024.0 * 1024.0)
		if sizeMB < 100 {
			// skip small files
			return nil
		}

		// Notify download start
		_, _ = m.Reply("📥 Gogram: Downloading…")

		// Determine a filename:
		// Prefer DocumentAttributeFilename if present, otherwise fallback to a generated name using doc.ID.
		base := ""
		for _, attr := range doc.Attributes {
			if f, ok := attr.(*telegram.DocumentAttributeFilename); ok {
				if f.FileName != "" {
					base = f.FileName
					break
				}
			}
		}
		if base == "" {
			// try DocumentObj.Name-like fields if available; otherwise fallback to ID-based name
			// DocumentObj may not expose a simple "Name" property, so we fallback safely:
			base = fmt.Sprintf("file_%d", doc.ID)
		}

		// sanitize/fallback
		if base == "" {
			base = fmt.Sprintf("file_%d", time.Now().Unix())
		}
		fileName := "gogram_" + filepath.Base(base)

		// Download
		start := time.Now()
		path, err := m.Download(&telegram.DownloadOptions{
			FileName: fileName,
		})
		if err != nil {
			_, _ = m.Reply("❌ Download failed: " + err.Error())
			return err
		}
		downloadTime := time.Since(start)

		_, _ = m.Reply("📤 Gogram: Uploading…")

		// Upload the downloaded file back to the same chat
		start = time.Now()
		_, err = client.SendMedia(m.ChatID(), path, &telegram.MediaOptions{
			Caption:   fmt.Sprintf("`%s`", base),
			ParseMode: "Markdown",
		})
		uploadTime := time.Since(start)
		if err != nil {
			_, _ = m.Reply("❌ Upload failed: " + err.Error())
			// cleanup file
			_ = os.Remove(path)
			return err
		}

		// cleanup local file
		_ = os.Remove(path)

		// Reply with summary
		_, _ = m.Reply(fmt.Sprintf(
			"✅ Gogram Result\n📦 Size: %.2f MB\n⬇️ Download: %s\n⬆️ Upload: %s",
			sizeMB,
			downloadTime.Truncate(time.Second),
			uploadTime.Truncate(time.Second),
		))

		return nil
	})

	fmt.Println("🚀 Gogram listener started")
	client.Idle()
}
