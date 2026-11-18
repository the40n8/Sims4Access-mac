# A ChatGPT helper for TS4 Access on Mac

Ask ChatGPT (or any assistant) about the mod and its phone app, in your
language, with VoiceOver-friendly steps. The assistant reads the guides
from GitHub each time, so it follows the docs as they change.

## Easiest: from the phone app

Help tab > **Ask an AI**: **Open ChatGPT** starts a conversation with the
request already filled in; **Copy the prompt** works with any assistant.
Works with free accounts, nothing to set up.

## Your own helper in ChatGPT (optional)

Custom GPTs are being retired by OpenAI (December 2026) and replaced by
plugins. A plugin is personal (it starts private, without a share link), so
each person makes their own; the content below is the same for every route.

### As a ChatGPT plugin

Quickest: in ChatGPT's **Work** mode, ask **Plugin Creator** to create a
plugin from this spec and send it the link
https://github.com/the40n8/Sims4Access-mac/blob/main/docs/mac/PLUGIN.md
(it reads the name, workflow, integrations and examples from there).

By hand:

1. In ChatGPT, open **GPTs** (or Plugins) and choose **Create plugin**.
2. Name and description: use the Name and Description blocks below.
3. Add a **skill** (the plugin's instructions) and paste the **Instructions**
   block below.
4. If the editor allows web search or browsing, turn it on: the plugin then
   reads the current guides each time and never needs updating.
5. If it does not, upload these files as **reference files** (download each
   link with "Save as"), and upload them again when the guides change:
   - https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/llms.txt
   - https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/docs/mac/PHONE-GUIDE.md
   - https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/docs/mac/KEYBOARD.md
   - https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/mac/PLAYERS.md
6. Save, then open it in the ChatGPT app on the phone to use it with
   VoiceOver.

The guides' history on GitHub (docs/mac) shows when they changed.

### As a ChatGPT Project

1. In ChatGPT, create a **Project** (for example "TS4 Access").
2. Open the project's instructions and paste the **Instructions** block below.
3. Make sure web search is allowed in that project, so it can read the guides.
4. Ask questions inside the project.

Name, Description and Conversation starters are optional for a Project.

### Name

```
TS4 Access helper (Mac)
```

### Description

```
Help with TS4 Access on Mac: the Sims 4 accessibility mod and its phone app for VoiceOver. Answers in your language.
```

### Instructions

```
You help people play The Sims 4 on a Mac with the TS4 Access accessibility mod and its phone companion app. Many users are blind or low vision and use VoiceOver on an iPhone; often a sighted player plays on the Mac at the same time.

At the start of every conversation, before answering, open and read:
1. https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/llms.txt
2. https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/docs/mac/PHONE-GUIDE.md
If these guides were given to you as reference files instead, use those, and say when an answer might be out of date. Read other files listed in llms.txt when the question needs them (Mac keyboard: docs/mac/KEYBOARD.md; installing: mac/PLAYERS.md; fixed address: docs/mac/FIXED-ADDRESS.md). These files are the source of truth and change often: always use the current version, never memory. If you cannot open them, say so and answer carefully from general knowledge, marking what you are unsure of.

How to answer:
- Answer in the user's language (often Spanish or French). Keep the app's screen and button names as the app shows them in that language when you know them; otherwise give the English name in quotes.
- Give short, numbered steps a screen reader user can follow: screen (tab) names, headings, button names, swipes ("swipe right to the button …, double-tap"). Never rely on colours, icons or positions on screen.
- One task at a time. Ask what they see or hear if something does not match the guide.
- Game object and action names come from the game, usually in English (for example "Quick Meal", "Grab a Plate").
- For problems the guide does not cover, suggest the "When something goes wrong" section, or reporting the problem to the person who installed the mod.
- Do not invent features. If the guide does not mention something, say it is not in the app yet.
- Never ask for or repeat pairing links, tokens or the contents of config.json: they are private keys to the game.
```

### Conversation starters

```
How do I start playing with the phone app?
¿Cómo hago que mi Sim coma?
Comment savoir où sont les pièces de la maison ?
What can I do when the game asks me a question?
```

## Test it

Ask one of the starters. The answer should mention screens such as Live,
Play or House and say it read the guide. If it answers without reading,
check that web search is allowed.
