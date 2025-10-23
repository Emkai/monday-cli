# Monday.com CLI

A powerful command-line tool for managing your Monday.com tasks with local caching, smart indexing, and efficient task operations.


## ✨ Features

- **🚀 Local Index System**: Tasks are numbered with user-friendly indices (1, 2, 3...) instead of long Monday.com IDs
- **💾 Smart Caching**: Tasks are cached locally for fast access and offline viewing
- **📊 Intelligent Sorting**: Tasks are sorted by status, priority, and type for optimal workflow
- **🎯 Quick Operations**: Show, edit, and manage tasks with simple commands

## Quick Start

1. **Build the tool:**
```bash
go mod tidy
go build -o mon .
```
> **Note for Windows users**: You can name the executable `mon.exe` for easier usage on Windows.

2. **Configure it:**
```bash
./mon config set-api-key <your-api-key>
./mon config set-board-id <your-board-id>
./mon config set-sprint-id <your-sprint-id>  # No implementation for sprints yet, still to come
```

3. **View your tasks:**
```bash
./mon tasks list
```

## 📋 Commands

### Task Management
- `./mon tasks list` - Show your cached tasks with local indices
- `./mon tasks fetch` - Fetch fresh tasks from Monday.com
- `./mon task show <index>` - Show details of a specific task
- `./mon task create <name> [flags]` - Create a new task
- `./mon task edit <index> [flags]` - Edit an existing task

### Configuration
- `./mon config show` - Display current configuration
- `./mon config set-api-key <key>` - Set your Monday.com API key
- `./mon config set-board-id <id>` - Set your board ID
- `./mon config set-sprint-id <id>` - Set your sprint ID (optional)

### User Management
- `./mon user info` - Show your user information

## 🎯 Task Creation & Editing

### Create Tasks with Flags
```bash
# Create a bug task with high priority
./mon task create "Fix login issue" -t b -p h -s p

# Create a feature task
./mon task create "Add dark mode" -t f -p m -s p
```

### Edit Tasks with Flags
```bash
# Update task status and priority
./mon task edit 1 -s d -p c

# Change task type and status
./mon task edit 5 -t f -s p
```

### Available Flags
- **Status**: `-s` or `-status` (done/d, in progress/p, stuck/s, waiting review/r, ready for testing/t, removed/rm)
- **Priority**: `-p` or `-priority` (critical/c, high/h, medium/m, low/l)
- **Type**: `-t` or `-type` (bug/b, feature/f, test/t, security/s, quality/q)

## 🏷️ Task Display Format

Tasks display as: `1. 🐛 [🔄 🔴] Fix login issue`

- **Number**: Local index for easy reference
- **Type Icon**: 🐛 Bug, ✨ Feature, 🧪 Test, 🔒 Security, 📈 Quality, 📝 Other
- **Status**: 🔄 In Progress, ✅ Done, 🚫 Blocked, 👀 Review, 🧪 Testing, 🗑️ Removed
- **Priority**: 🔴 Critical, 🟡 High, 🔵 Medium, 🟢 Low, ⚪ Default

## 🔧 Getting Credentials

- **API Key**: Get from URL: `https://example.monday.com/apps/manage/tokens` (replace "example" with your team name)
- **Board ID**: Found in your board URL: `https://example.monday.com/boards/1234567890`
- **Sprint ID**: Optional, for filtering tasks by sprint

## 🛠️ Development

Built with Go, featuring:
- **GraphQL Integration**: Direct Monday.com API integration
- **Efficient Caching**: JSON-based local storage
- **Smart Parsing**: Robust command-line argument parsing
- **Error Handling**: Comprehensive error messages and validation
