# Monday.com CLI

A command-line tool for viewing your Monday.com tasks.

## Quick Start

1. **Build the tool:**
```bash
go build -o monday-cli .
```

2. **Configure it:**
```bash
monday-cli config set-api-key <your-api-key>
monday-cli config set-owner-email your.email@company.com
monday-cli config set-board-id <your-board-id>
```

3. **View your tasks:**
```bash
monday-cli tasks
```

## Commands

- `monday-cli tasks` - Show your assigned tasks
- `monday-cli config show` - Display current configuration
- `monday-cli help` - Show help

## Getting Credentials

- **API Key**: Get from URL: `https://example.monday.com/apps/manage/tokens` (replace "example" with your team name)
- **Board ID**: Found in your board URL: `https://example.monday.com/boards/1234567890`

## Task Format
Tasks display as: `📝 [🔄 ⚪] Task Name`
- First emoji: Type (📝 Other/Default, 🐛 Bug, ✨ Feature, 🧪 Test, 🔒 Security, 📈 Quality)
- Status: 🔄 In Progress, ✅ Done, 🚫 Blocked, 👀 Review, 🧪 Testing/Not Started, 🗑️ Removed, 📋 Default
- Priority: 🔴 Critical, 🟡 High, 🔵 Medium, 🟢 Low, ⚪ Default
