---
name: commit
description: Commit Skill with conventional commits and gitmoji support
allowed-tools: Bash(git *)
---

# Commit Skill

Create git commits following Conventional Commits specification with gitmoji.

## Format

```
<gitmoji> <type>(<scope>): <subject>

[optional body]

[optional footer(s)]
```

## Gitmoji Reference

| Emoji | Code | Type | Description |
|-------|------|------|-------------|
| ✨ | `:sparkles:` | feat | New feature |
| 🐛 | `:bug:` | fix | Bug fix |
| 📝 | `:memo:` | docs | Documentation |
| 💄 | `:lipstick:` | style | UI/style changes |
| ♻️ | `:recycle:` | refactor | Code refactoring |
| ⚡️ | `:zap:` | perf | Performance improvement |
| ✅ | `:white_check_mark:` | test | Tests |
| 🔧 | `:wrench:` | chore | Configuration/tooling |
| 🏗️ | `:building_construction:` | build | Build system changes |
| 👷 | `:construction_worker:` | ci | CI configuration |
| 🔥 | `:fire:` | remove | Remove code/files |
| 🚀 | `:rocket:` | deploy | Deployment |
| 🔒 | `:lock:` | security | Security fix |
| ⬆️ | `:arrow_up:` | deps | Upgrade dependencies |
| ⬇️ | `:arrow_down:` | deps | Downgrade dependencies |
| 🎨 | `:art:` | style | Improve structure/format |
| 💚 | `:green_heart:` | ci | Fix CI build |
| 📦 | `:package:` | build | Update compiled files |
| 🚧 | `:construction:` | wip | Work in progress |
| 🔀 | `:twisted_rightwards_arrows:` | merge | Merge branches |
| ⏪ | `:rewind:` | revert | Revert changes |
| 🏷️ | `:label:` | types | Types (TypeScript) |
| 🩹 | `:adhesive_bandage:` | fix | Simple fix for non-critical issue |
| 🧪 | `:test_tube:` | test | Add failing test |
| 💡 | `:bulb:` | docs | Add comments in code |
| 🍱 | `:bento:` | assets | Add/update assets |
| ♿️ | `:wheelchair:` | a11y | Accessibility |
| 📱 | `:iphone:` | responsive | Responsive design |
| 🗃️ | `:card_file_box:` | db | Database changes |
| 🔊 | `:loud_sound:` | logs | Add logs |
| 🔇 | `:mute:` | logs | Remove logs |

## Examples

```bash
# New feature
✨ feat(auth): add login with Google OAuth

# Bug fix with scope
🐛 fix(api): resolve null pointer in user service

# Test updates
✅ test: update TaskCardBody gap assertion and colors snapshot

# Documentation
📝 docs(readme): add installation instructions

# Refactoring
♻️ refactor(utils): simplify date formatting logic

# Performance
⚡️ perf(queries): optimize database queries with indexing

# Dependencies
⬆️ deps: upgrade react-native to 0.73

# Breaking change (add ! after type)
✨ feat(api)!: change authentication flow

BREAKING CHANGE: API now requires Bearer token
```

## Workflow

1. Check `git status` and `git diff` to understand changes
2. Stage changes with `git add`
3. Determine the appropriate type and gitmoji
4. Write a concise, imperative subject (max 72 chars)
5. Add body if changes need explanation
6. Commit with the formatted message
