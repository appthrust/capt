# 1. Keep a Changelogの形式の採用

Date: 2024-11-13

## Status

Accepted

## Context

プロジェクトの変更履歴を追跡し、ユーザーや開発者に変更内容を明確に伝える必要があります。変更履歴の記録方法には様々な形式が存在し、以下の要件を満たす必要があります：

- 人間が読みやすい形式であること
- 変更の種類（新機能、バグ修正など）が明確に区別できること
- リリースバージョンと日付が明確であること
- 自動化ツールでの解析が容易であること

## Decision

[Keep a Changelog](https://keepachangelog.com/)の形式を採用します。具体的には：

1. CHANGELOGは`CHANGELOG.md`として記録
2. 以下のセクションを使用して変更を分類：
   - Added: 新機能
   - Changed: 既存機能の変更
   - Deprecated: 将来削除される機能
   - Removed: 削除された機能
   - Fixed: バグ修正
   - Security: セキュリティ関連の修正

3. 各リリースは以下の形式で記録：
```markdown
## [version] - YYYY-MM-DD

### Added
- 新機能の説明

### Fixed
- バグ修正の説明
```

## Consequences

### Positive

- 標準化された形式により、変更履歴が一貫性を持つ
- 変更の種類が明確に区別され、ユーザーが必要な情報を見つけやすい
- セマンティックバージョニングとの親和性が高い
- 多くのツールやプラットフォームでサポートされている

### Negative

- 各変更を適切なカテゴリに分類する必要がある
- すべての開発者がフォーマットを理解し、従う必要がある

## References

- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
