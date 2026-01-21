import time
import os
from telethon import TelegramClient, events

API_ID = int(os.environ["API_ID"])
API_HASH = os.environ["API_HASH"]
BOT_TOKEN = os.environ["BOT_TOKEN"]

client = TelegramClient("telethon_session", API_ID, API_HASH)

@client.on(events.NewMessage(incoming=True))
async def handler(event):
    if not event.file:
        return

    size_mb = event.file.size / (1024 * 1024)
    if size_mb < 100:  # ignore small files
        return

    await event.reply("📥 Telethon: Downloading...")

    file_name = f"telethon_{event.file.name or 'file.bin'}"

    # ⬇️ Download
    t1 = time.time()
    await client.download_media(event.message, file_name)
    d_time = time.time() - t1

    await event.reply("📤 Telethon: Uploading...")

    # ⬆️ Upload
    t2 = time.time()
    await client.send_file(event.chat_id, file_name)
    u_time = time.time() - t2

    await event.reply(
        f"✅ **Telethon Result**\n"
        f"📦 Size: `{size_mb:.2f} MB`\n"
        f"⬇️ Download: `{d_time:.2f}s`\n"
        f"⬆️ Upload: `{u_time:.2f}s`"
    )

client.start(bot_token=BOT_TOKEN)
client.run_until_disconnected()
